# Building a production-ready LLM Proxy Server with Go: From Zero to Docker Deployment

In the era of AI, managing connections to Large Language Models (LLMs) like OpenAI, Azure, or Anthropic can quickly become messy. Security keys get scattered, rate limits are hit unexpectedly, and changing providers involves code refactoring.

Today, I’ll walk you through building a robust **LLM Proxy Server** in Go (Golang) that solves these problems. We’ll design it to be production-ready with **Rate Limiting**, **Virtual Keys**, and support for multiple databases (SQLite, Postgres, MongoDB), finally deploying it with Docker.

---

## The Concept

We want a middle layer that sits between your users/apps and the LLM providers.
**Core Features:**
1.  **Connection Abstraction**: Store real API keys securely in the server.
2.  **Virtual Keys**: Issue new keys to your internal teams/users. These keys map to real connections.
3.  **Rate Limiting**: Control how many TPS (Transactions Per Second) or Tokens each virtual key can consume.
4.  **Database Agnostic**: Use SQLite for simple usage, or switch to Postgres/MongoDB for scale.

## The Stack

We choose **Go** for its performance and concurrency model, which is perfect for proxy servers.
-   **Framework**: [Gin](https://github.com/gin-gonic/gin) for high-performance HTTP routing.
-   **CLI**: [Cobra](https://github.com/spf13/cobra) for a great command-line interface.
-   **ORM**: [GORM](https://gorm.io/) and [Mongo Driver](https://go.mongodb.org/mongo-driver/v2).
-   **Limiter**: Token bucket algorithm via `golang.org/x/time/rate`.

## Architecture Highlights

### 1. Unified Database Layer
We designed a `DB` interface that abstracts the underlying storage. This allows us to switch from SQLite to MongoDB with just a config flag.

```go
type DB interface {
    SaveConnection(ctx context.Context, conn *models.Connection) error
    GetConnection(ctx context.Context, id string) (*models.Connection, error)
    // ...
}
```

### 2. Multi-Dimensional Rate Limiting
Most proxies only limit requests per second (TPS). But in the LLM world, **tokens** are the real currency. Our implementation limits both:

```go
// internal/ratelimit/ratelimit.go
type Limiter struct {
    tpsLimiter   *rate.Limiter // For API calls limit
    tokenLimiter *rate.Limiter // For Token consumption limit
}
```
This ensures a user can't drain your budget even if they send few requests but with massive contexts.

### 3. The Proxy Logic
The proxy intercepts the request, validates the `Virtual Key`, enforces limits, and then delegates the call to the real provider using Go's `http.Client`. It’s essentially a transparent pipe that adds security and control.

## Deployment with Docker

We use a Multi-Stage Dockerfile to keep our image tiny.

```dockerfile
# Build Stage
FROM golang:1.24-alpine AS builder
# ... compile binary ...

# Final Stage
FROM alpine:latest
COPY --from=builder /app/llm-proxy .
ENTRYPOINT ["./llm-proxy"]
```
This results in a lightweight container (~20MB compressed) that starts instantly.

## How to Run It

clone the repo and fire it up with Docker Compose:

```bash
docker compose up -d
```

Add your real OpenAI key securely via CLI inside the container (or locally):
```bash
./llm-proxy connection add --name "OpenAI-Prod" --provider "openai" --endpoint "..." --api-key "sk-..."
```

Create a limited virtual key for your intern:
```bash
./llm-proxy vkey add --name "Intern-Key" --conn-id "..." --tps 2 --tokens 1000
```

Now, your intern uses `http://localhost:8132` with their virtual key. If they spam requests, your proxy blocks them, saving your real API quota.

## Conclusion

This project demonstrates how Go's standard library combined with powerful tools like Gin and GORM can create enterprise-grade tools with minimal boilerplate. By adding abstraction layers like Virtual Keys, you gain visibility and control over your AI infrastructure.

---
*Check out the full source code in the repository.*
