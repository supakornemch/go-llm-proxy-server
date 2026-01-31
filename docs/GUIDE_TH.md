# คู่มือการใช้งานและโครงสร้างระบบ LLM Proxy Server (ฉบับละเอียด)

เอกสารนี้อธิบายสถาปัตยกรรมภายในของระบบ LLM Proxy Server พร้อมคำแนะนำวิธีการตั้งค่า Connection ไปยัง AI Provider เจ้าดังต่างๆ อย่างละเอียด

---

## 🏗 System Architecture (โครงสร้างระบบ)

แผนภาพด้านล่างแสดงการทำงานของระบบเมื่อ Client (เช่น Python Script, cURL) ส่ง Request เข้ามายัง Proxy:

```mermaid
%%{init: {'theme':'base', 'themeVariables': { 'primaryColor':'#1f77b4', 'primaryBorderColor':'#004a9e', 'lineColor':'#666', 'secondColor':'#2ca02c', 'tertiaryColor':'#ff7f0e'}, 'flowchart': {'useMaxWidth': true, 'padding': '20', 'fontSize': '14'}}}%%
flowchart TD
    Client["👤 Client App / SDK<br/>(Python, Node.js, cURL)"]
    
    Client -->|"📤 HTTP Request<br/>(Auth: Bearer Virtual-Key)"| Proxy["🔐 GO Proxy Server<br/><br/>Port 8132"]
    
    subgraph ProxyLogic["<b>⚙️ Proxy Server Logic</b>"]
        Proxy -->|"1️⃣ Validate Key"| DB[("💾 Database<br/>MongoDB/SQL<br/><br/>Stores Keys,<br/>Assignments")]
        DB -->|"✅ Return Virtual Key Data"| Proxy
        
        Proxy -->|"2️⃣ Check Assignment"| Logic["🔀 Routing Logic<br/><br/>Maps Virtual Model<br/>→ Real Model"]
        Logic -->|"📋 Lookup Config"| DB
        
        Proxy -->|"3️⃣ Rate Limiting"| RateLimiter["⏱️ Token Bucket<br/>Limiter<br/><br/>TPS Control"]
        RateLimiter -->|"✅ OK"| Adapter["🔄 Protocol Adapter<br/><br/>Transform to Provider<br/>Format"]
        RateLimiter -->|"⛔ Exceeded"| Reject["⚠️ 429<br/>Too Many<br/>Requests"]
    end

    subgraph AdapterLogic["<b>🌐 Adapter Logic</b>"]
        Adapter -->|"Detect Provider"| Azure{"☁️ Azure<br/>OpenAI?"}
        Adapter -->|"Detect Provider"| Google{"🔍 Google<br/>Vertex/Studio?"}
        Adapter -->|"Detect Provider"| Standard{"📌 OpenAI/<br/>AWS?"}

        Azure -->|"✏️ Inject: api-key<br/>Rewrite: URL + version"| AzureEP["☁️ Azure OpenAI<br/>Endpoint<br/><br/>https://xxx.openai.azure.com"]
        Google -->|"✏️ Inject: x-goog-api-key<br/>Strip: Bearer (if API key)"| GoogleEP["🔍 Google Vertex/Studio<br/>Endpoint<br/><br/>aiplatform.googleapis.com"]
        Standard -->|"✏️ Inject: Bearer Token"| StandardEP["📌 OpenAI / AWS Bedrock<br/>Endpoint<br/><br/>api.openai.com"]
    end

    AzureEP -->|"📥 Response"| Client
    GoogleEP -->|"📥 Response"| Client
    StandardEP -->|"📥 Response"| Client
    Reject -->|"❌ Error"| Client
    
    style Client fill:#e1f5ff,stroke:#01579b,stroke-width:2px,color:#000
    style Proxy fill:#fff3e0,stroke:#e65100,stroke-width:3px,color:#000,font-weight:bold
    style DB fill:#f3e5f5,stroke:#4a148c,stroke-width:2px,color:#000
    style Logic fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px,color:#000
    style RateLimiter fill:#fce4ec,stroke:#880e4f,stroke-width:2px,color:#000
    style Adapter fill:#fff9c4,stroke:#f57f17,stroke-width:2px,color:#000
    style Reject fill:#ffebee,stroke:#b71c1c,stroke-width:2px,color:#000
    style Azure fill:#e3f2fd,stroke:#1565c0,stroke-width:2px,color:#000
    style Google fill:#f1f8e9,stroke:#558b2f,stroke-width:2px,color:#000
    style Standard fill:#ede7f6,stroke:#512da8,stroke-width:2px,color:#000
    style AzureEP fill:#bbdefb,stroke:#0d47a1,stroke-width:2px,color:#000
    style GoogleEP fill:#c8e6c9,stroke:#2e7d32,stroke-width:2px,color:#000
    style StandardEP fill:#d1c4e9,stroke:#3949ab,stroke-width:2px,color:#000
    style ProxyLogic fill:#fff8e1,stroke:#f57c00,stroke-width:2px
    style AdapterLogic fill:#f1f8e9,stroke:#689f38,stroke-width:2px
```

### คำอธิบายส่วนประกอบหลัก (Components)
1.  **Proxy Handler (`internal/proxy`)**: ด่านหน้าสำหรับรับ HTTP Request ทำหน้าที่:
    -   แกะ `Authorization` Header เพื่อหา Virtual Key
    -   อ่าน Body หรือ URL Path เพื่อหาว่า User ต้องการเรียก Model อะไร (เช่น `gpt-4`, `gemini-1.5`)
2.  **Database (`internal/db`)**: เก็บข้อมูล 4 ส่วนหลัก:
    -   `Connections`: เก็บ Credential จริงของ Provider (เช่น OpenAI API Key) **(ถูกเข้ารหัสเก็บไว้)**
    -   `ProviderModels`: เก็บชื่อ Model จริงในระบบ Provider (เช่น `gemini-1.5-flash-001`)
    -   `VirtualKeys`: กุญแจที่ Proxy สร้างขึ้นแจกจ่ายให้ Client
    -   `Assignments`: ตารางจับคู่ว่า Virtual Key นี้ มีสิทธิ์ใช้ Model ไหนได้บ้าง
3.  **Rate Limiter (`internal/ratelimit`)**: คอยนับจำนวน Request และ Token ที่ถูกใช้ไปในแต่ละวินาที ถ้าเกินกำหนดจะตีกลับทันที
4.  **Protocol Adapter**: (สำคัญมาก) ทำหน้าที่แปลง Request ให้เข้ากับมาตรฐานของแต่ละค่าย เช่น:
    -   **Azure**: ต้องเติม `?api-version=...` และใช้ Header `api-key`
    -   **Google Vertex/Gemini**: ต้องสลับระหว่าง `x-goog-api-key` หรือ `Authorization: Bearer` ตามชนิดของ Key ที่ใช้

---

## 🛠 วิธีการสร้าง Connection ไปยัง Provider ต่างๆ

ก่อนเริ่มใช้งาน ต้องรัน Server ด้วยคำสั่ง:
```bash
# รันผ่าน Docker Compose
docker compose up -d

# หรือรัน Binary
./llm-proxy serve
```

### 1. OpenAI (Standard)
OpenAI เป็นมาตรฐานกลางที่ง่ายที่สุด

**ข้อมูลที่ต้องเตรียม:**
-   **API Key**: `sk-...`
-   **Endpoint**: `https://api.openai.com`

**คำสั่ง:**
```bash
# 1. สร้าง Connection
./llm-proxy connection add \
  --name "OpenAI-Main" \
  --provider "openai" \
  --endpoint "https://api.openai.com" \
  --api-key "sk-proj-YourKey..."

# (สมมติได้ ID: conn-123)

# 2. เพิ่ม Model เข้าไปใน Connection นี้
./llm-proxy model add \
  --conn-id "conn-123" \
  --name "gpt-4-turbo" \
  --remote "gpt-4-turbo-preview"
```

### 2. Azure OpenAI Service
Azure มีรูปแบบ URL ที่ซับซ้อนกว่า โดยมักจะอยู่ในรูป `https://{resource}.openai.azure.com/` หรือรูปแบบ Foundry

**ข้อมูลที่ต้องเตรียม:**
-   **API Key**: Key จาก Azure Portal
-   **Endpoint**: URL หน้าตาประมาณ `https://my-resource.openai.azure.com` หรือ Foundry URL

**คำสั่ง:**
```bash
# 1. สร้าง Connection
./llm-proxy connection add \
  --name "Azure-Corp" \
  --provider "azure" \
  --endpoint "https://my-company.openai.azure.com" \
  --api-key "your-azure-key"

# (สมมติได้ ID: conn-456)

# 2. เพิ่ม Model (Deployment Name สำคัญมากใน Azure)
./llm-proxy model add \
  --conn-id "conn-456" \
  --name "gpt-4o" \
  --remote "gpt-4o" \
  --deployment "deployment-name-in-azure"
```

> **Note:** Proxy จะเติม `?api-version=2024-05-01-preview` ให้เองอัตโนมัติหากไม่ได้ระบุมา

### 3. Google Gemini (AI Studio)
สำหรับผู้ใช้ Google AI Studio (API Key ปกติ)

**ข้อมูลที่ต้องเตรียม:**
-   **API Key**: Key จาก aistudio.google.com
-   **Endpoint**: `https://generativelanguage.googleapis.com`

**คำสั่ง:**
```bash
# 1. สร้าง Connection
./llm-proxy connection add \
  --name "Gemini-AIStudio" \
  --provider "google" \
  --endpoint "https://generativelanguage.googleapis.com" \
  --api-key "AIzaSy..."

# (สมมติได้ ID: conn-789)

# 2. เพิ่ม Model
./llm-proxy model add \
  --conn-id "conn-789" \
  --name "gemini-1.5-flash" \
  --remote "gemini-1.5-flash"
```

### 4. Google Vertex AI (Enterprise)
สำหรับองค์กรที่ใช้ Vertex AI บน Google Cloud

**ข้อมูลที่ต้องเตรียม:**
-   **API Key**: Service Account Key หรือ API Key (ขึ้นต้นด้วย `AQ.`) หรือ OAuth Token
-   **Endpoint**: `https://aiplatform.googleapis.com`

**คำสั่ง:**
```bash
# 1. สร้าง Connection
./llm-proxy connection add \
  --name "Vertex-Prod" \
  --provider "google" \
  --endpoint "https://aiplatform.googleapis.com" \
  --api-key "AQ.Ab8..." # หรือ OAuth Token

# (สมมติได้ ID: conn-999)

# 2. เพิ่ม Model
./llm-proxy model add \
  --conn-id "conn-999" \
  --name "gemini-3-flash" \
  --remote "gemini-3-flash-preview"
```

---

## 🔑 การใช้งานฝั่ง Client (Usage)

เมื่อตั้งค่า connection เสร็จแล้ว ผู้ใช้ฝั่ง Client ต้องทำ 2 ขั้นตอนนี้:

1.  **สร้าง Virtual Key** (Admin ทำให้):
    ```bash
    ./llm-proxy vkey add --name "Frontend-App" --key "vk-front-1234"
    ```
2.  **กำหนดสิทธิ์ (Assign)** ว่า Key นี้ใช้ Model ไหนได้บ้าง:
    ```bash
    # ผูก Virtual Key เข้ากับ Model ID ที่เราสร้างไว้ข้างบน
    ./llm-proxy assign \
      --vkey-id "vkey-id..." \
      --model-id "model-id..." \
      --alias "gpt-4" \
      --tps 50 # ยิงได้ 50 ครั้งต่อวินาที
    ```

### ตัวอย่าง Code (Python)

#### 1. การใช้งานผ่าน OpenAI SDK (มาตรฐาน)
หากคุณใช้ Model อย่าง GPT-4 หรือ Gemini ที่ Config เป็น OpenAI-Compatible:

```python
from openai import OpenAI

client = OpenAI(
    api_key="vk-front-1234",          # ใช้ Virtual Key ที่ได้จาก Proxy
    base_url="http://localhost:8132/v1"  # ชี้มาที่ Proxy Server (เติม /v1)
)

response = client.chat.completions.create(
    model="gpt-4", # ใช้ชื่อ Alias ที่ตั้งไว้ตอน Assign
    messages=[{"role": "user", "content": "สวัสดี!"}]
)

print(f"OpenAI Output: {response.choices[0].message.content}")
```

#### 2. การใช้งานผ่าน Google Generative AI SDK (Native)
หากต้องการใช้ฟีเจอร์เฉพาะของ Gemini เช่น **Thinking Config** ของ Gemini 2.0/3.0:

```python
import google.generativeai as genai

# ตั้งค่าให้ชี้มาที่ Proxy
genai.configure(
    api_key="vk-front-1234",
    client_options={
        "api_endpoint": "http://localhost:8132" # ชี้มาที่ Proxy
    },
    transport="rest" # สำคัญ: ต้องใช้ REST transport เท่านั้น
)

model = genai.GenerativeModel("gemini-3-flash")

# ตัวอย่างการใช้ Thinking Config
response = model.generate_content(
    "อธิบายเรื่อง Quantum Physics สั้นๆ",
    generation_config={
        "thinking_config": {"include_thoughts": True}
    }
)

print(f"Gemini Output: {response.text}")
```

#### 3. การใช้งานผ่าน Azure OpenAI SDK
สำหรับองค์กรที่คุ้นเคยกับรูปแบบของ Azure SDK:

```python
from openai import AzureOpenAI

client = AzureOpenAI(
    api_key="vk-front-1234",
    api_version="2024-05-01-preview", # หรือ version อื่นๆ
    azure_endpoint="http://localhost:8132" # ชี้มาที่ Proxy
)

# หมายเหตุ: 'model' ในที่นี้คือ Deployment Name หรือ Alias ที่ตั้งใน Proxy
response = client.chat.completions.create(
    model="azure-gpt-4o",
    messages=[{"role": "user", "content": "Hello Azure!"}]
)

print(f"Azure Output: {response.choices[0].message.content}")
```

#### 4. การใช้งานผ่าน LangChain
LangChain นิยมมากในการสร้าง LLM App:

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    model="gemini-3-flash",
    openai_api_key="vk-front-1234",
    openai_api_base="http://localhost:8132/v1", # ชี้มาที่ Proxy
    temperature=0
)

response = llm.invoke("เล่านิทานให้ฟังหน่อย")
print(f"LangChain Output: {response.content}")
```
