FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o llm-proxy main.go

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache sqlite
COPY --from=builder /app/llm-proxy .

EXPOSE 8132
ENTRYPOINT ["./llm-proxy", "serve"]
