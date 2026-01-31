# Go LLM Proxy Server

A lightweight, reliable LLM proxy server written in Go. Manage multiple LLM providers behind a unified endpoint with virtual keys and rate limiting.

## Features

- **Multi-Provider Support**: Connect to OpenAI, Azure OpenAI, and Google Gemini (via OpenAI-compatible endpoint).
- **Smart Protocol Adaptation**: Automatically handles authentication schemes (Bearer, API Key headers, or Query parameters) based on the provider.
- **Virtual Keys**: Issue specific keys to clients/projects without exposing master API keys.
- **Rate Limiting**: Control usage with Requests Per Second (TPS) and Token limits.
- **Unified Interface**: Supports SQL (SQLite, Postgres, MSSQL) and NoSQL (MongoDB).
- **Cloud Ready**: Easily deployable to Azure App Service, Heroku, or Docker.

## Documentation

- [English Article (Concept & Design)](docs/ARTICLE_EN.md)
- [Thai Article (บทความภาษาไทย)](docs/ARTICLE_TH.md)

## Quick Start (Docker)

1. Clone the repository.
2. Create `.env` from `.env.example`.
3. Run with Docker Compose:
   ```bash
   docker compose up --build -d
   ```

## Usage Examples

### 1. LangChain (Python)

The proxy is fully compatible with OpenAI SDKs. You can access **Google Gemini** models using the standard `ChatOpenAI` class!

```python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    model="gemini-1.5-flash",       # Use the alias assigned in the proxy
    api_key="sk-proxy-default-key", # Your Virtual Key
    base_url="http://localhost:8132/v1"
)

print(llm.invoke("Hello from Gemini via Proxy!").content)
```

### 2. Google GenAI SDK (Native & Thinking Config)

The proxy also supports the native Google GenAI SDK and advanced features like **Thinking Config** (Gemini 2.0).

```python
from google import genai
from google.genai import types

client = genai.Client(
    api_key="sk-proxy-default-key", # Your Virtual Key
    request_options={"base_url": "http://localhost:8132"}
)

# You can use advanced Gemini features through the proxy passthrough
config = types.GenerateContentConfig(
    thinking_config=types.ThinkingConfig(thinking_level="HIGH"),
)

response = client.models.generate_content(
    model="gemini-2.0-flash-thinking-preview", 
    contents="Explain quantum computing.",
    config=config
)
print(response.text)
```

### 3. cURL

```bash
curl -X POST http://localhost:8132/v1/chat/completions \
  -H "Authorization: Bearer sk-proxy-default-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### 3. Google Gemini Setup

To add Google Gemini as a provider, use the new OpenAI-compatible endpoint:

- **Provider**: `google`
- **Endpoint**: `https://generativelanguage.googleapis.com/v1beta/openai` (Note: specific path for OpenAI compatibility)
- **API Key**: Your Google AI Studio Key

## Cloud Deployment (e.g., Azure App Service)

To deploy without manually running CLI commands, the server supports **Auto-Seeding**. Set the following Environment Variables in your App Service configuration:
- `DB_TYPE`: `sqlite` (or `postgres`)
- `MASTER_CONN_NAME`: `Production-LLM`
- `MASTER_CONN_PROVIDER`: `openai`
- `MASTER_CONN_ENDPOINT`: `https://api.openai.com`
- `MASTER_CONN_API_KEY`: `your-real-api-key`
- `MASTER_CONN_MODEL`: `gpt-4o`
- `MASTER_VKEY_NAME`: `Default-Access`
- `MASTER_VKEY_KEY`: `your-proxy-access-key`

The server will automatically register these in the database on its first run if they don't exist.

## CLI Usage

Manage connections manually:
```bash
./llm-proxy connection add --name "Alpha" --provider "openai" --endpoint "..." --api-key "..."
./llm-proxy vkey add --name "TestKey" --conn-id "UUID" --key "sk-proxy-123"
```

## License

MIT
