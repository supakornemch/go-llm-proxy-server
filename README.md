# Go LLM Proxy Server

A lightweight, reliable LLM proxy server written in Go. Manage multiple LLM providers behind a unified endpoint with virtual keys and rate limiting.

## Features

- **Multi-Provider Support**: Connect to OpenAI, Azure OpenAI, and Google Gemini (Native & AI Platform).
- **Smart Protocol Adaptation**: Automatically handles authentication schemes (Bearer, API Key headers, or Query parameters) based on the provider.
- **Virtual Keys**: Issue specific keys to clients/projects without exposing master API keys.
- **Rate Limiting**: Control usage with Requests Per Second (TPS) and Token limits.
- **Unified Interface**: Supports SQL (SQLite, Postgres, MSSQL) and NoSQL (MongoDB).
- **Cloud Ready**: Easily deployable to Azure App Service, Heroku, or Docker.

## Quick Start (Docker)

1. Clone the repository.
2. Create `.env` from `.env.example`.
3. Run with Docker Compose:
   ```bash
   docker compose up --build -d
   ```

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
