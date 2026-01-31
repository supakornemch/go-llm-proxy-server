# Examples - LLM Proxy SDK Integration

ไฟล์ตัวอย่าง Python สำหรับใช้งาน LLM Proxy กับ SDK ต่าง ๆ

## 📚 ตัวอย่างทั้งหมด

| File | Provider | Description |
|------|----------|-------------|
| `example_openai.py` | OpenAI | ใช้ OpenAI SDK ผ่าน Proxy |
| `example_azure.py` | Azure OpenAI | ใช้ Azure OpenAI SDK ผ่าน Proxy |
| `example_google_gemini.py` | Google Gemini | ใช้ Google Generative AI SDK ผ่าน Proxy |
| `example_google_vertex_http.py` | Google Vertex AI | ส่ง Native JSON ไปยัง Proxy (รองรับ Thinking Config) |
| `example_raw_http.py` | Universal | ส่ง Raw HTTP Request ไปยัง Proxy (รองรับทุก Provider) |
| `example_bedrock.py` | AWS Bedrock | ใช้ boto3 สำหรับ Bedrock ผ่าน Proxy |

## 🚀 Quick Start

### 1. ติดตั้ง Dependencies

```bash
python3 examples/setup.py
```

หรือติดตั้งด้วยตัวเองแบบนี้:
```bash
pip install openai azure-openai google-generativeai boto3 requests
```

### 2. เปิด Proxy Server

```bash
./llm-proxy serve
```

### 3. ตั้งค่า Connection และ Virtual Key

ดู [คู่มือ Thai](../docs/GUIDE_TH.md) สำหรับวิธีการตั้งค่า

### 4. รันตัวอย่างที่ต้องการ

```bash
# OpenAI SDK
python3 examples/example_openai.py

# Azure OpenAI SDK
python3 examples/example_azure.py

# Google Gemini SDK
python3 examples/example_google_gemini.py

# Google Vertex AI with Native JSON (Thinking Config)
python3 examples/example_google_vertex_http.py

# Raw HTTP Requests (Universal)
python3 examples/example_raw_http.py

# AWS Bedrock (requires AWS setup)
python3 examples/example_bedrock.py
```

## ⚙️ Configuration

แต่ละไฟล์มี placeholder สำหรับแก้ไข:
- `vk-xxx`: Virtual Key (ขึ้นต้นด้วย `vk-`)
- `localhost:8132`: Proxy Server URL (เปลี่ยนตามที่ตั้งค่า)
- `model-name`: Alias ที่ตั้งไว้ตอน Assign

## 💡 Tips

- ตรวจสอบว่า Proxy Server รันอยู่ก่อนรัน Example
- ตรวจสอบ Virtual Key ถูกต้อง
- ดูข้อความ Error เพื่อแก้ไขปัญหา

## 📚 More Information

ดู [GUIDE_TH.md](../docs/GUIDE_TH.md) สำหรับวิธีการตั้งค่า Connection, Model, และ Virtual Key อย่างละเอียด
