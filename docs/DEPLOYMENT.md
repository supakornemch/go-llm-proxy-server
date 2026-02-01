# 🚀 Deployment Guide

## Docker (Recommended)

The project includes a multi-stage `Dockerfile` that handles CGO requirements for SQLite.

### 1. Configure Environment
Create a `.env` file:
```env
PORT=8080
DB_TYPE=sqlite
DATABASE_URL=sqlite.db
MASTER_KEY=your-admin-bypass-key
```

### 2. Build and Run
```bash
docker compose up --build -d
```

## Manual Deployment (Linux/macOS)

### 1. Install Go 1.21+
Ensure you have CGO enabled if using SQLite:
```bash
export CGO_ENABLED=1
go build -o llm-proxy main.go
```

### 2. Run as a Service
You can use `systemd` or `pm2` to manage the process.

## Environment Variables Reference

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Hub listening port | `8080` |
| `DB_TYPE` | `sqlite`, `postgres`, or `mongodb` | `sqlite` |
| `DATABASE_URL` | Connection string or file path | `sqlite.db` |
| `MASTER_KEY` | Admin bypass key | (none) |
| `LOG_LEVEL` | `debug`, `info`, `warn`, `error` | `info` |

## Database Seeding (Auto-Configuration)

The server supports auto-seeding via environment variables for easy cloud deployment (Azure, Heroku, etc.):

- `MASTER_CONN_NAME`: Name for the default connection.
- `MASTER_CONN_PROVIDER`: Provider type (`openai`, `google`, `azure`, `aws`).
- `MASTER_CONN_ENDPOINT`: Provider endpoint.
- `MASTER_CONN_API_KEY`: Actual provider API Key.
- `MASTER_CONN_MODEL`: Default model to register.
- `MASTER_VKEY_NAME`: Default virtual key name.
- `MASTER_VKEY_KEY`: Default virtual key value.
