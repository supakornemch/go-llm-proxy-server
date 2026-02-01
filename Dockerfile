FROM golang:1.24-alpine AS builder
# Install build dependencies for CGO
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Enable CGO for sqlite3 support
RUN CGO_ENABLED=1 GOOS=linux go build -o llm-proxy main.go

FROM alpine:latest
# Install sqlite library for runtime
RUN apk add --no-cache sqlite-libs

WORKDIR /app
COPY --from=builder /app/llm-proxy .

EXPOSE 8132
ENTRYPOINT ["./llm-proxy", "serve"]
