#!/bin/bash

# =================================================================================================
# LLM Proxy Server - Comprehensive Smoke Test (v2.0)
# =================================================================================================
# This script performs an end-to-end test of the LLM Proxy Server.
# It handles lifecycle management: Build -> Clean DB -> Seed Data -> Start Server -> Test -> Cleanup
# 
# Supported Providers in this test:
# 1. OpenAI (Official API)
# 2. Azure OpenAI (Foundry/AI Services)
# 3. Google Gemini (via OpenAI-Compatible Endpoint)
# 4. AWS Bedrock (Claude)
# =================================================================================================

# ----------------- Configuration -----------------
PORT=8132
VKEY="sk-smoke-test-$(date +%s)" # A fresh virtual key for this run
DB_TYPE=mongodb
DB_DSN="mongodb://root:examplepassword@localhost:27017/llm_proxy?authSource=admin"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}🚀 Starting LLM Proxy Smoke Test Setup...${NC}"

# ----------------- 0. Cleanup Previous State -----------------
echo -e "${YELLOW}🧹 Cleaning up old processes...${NC}"
pkill llm-proxy || true

# ----------------- 1. Build & Environment -----------------
echo -e "${YELLOW}🔨 Building project...${NC}"
go build -o llm-proxy main.go
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Build failed!${NC}"
    exit 1
fi

export DB_TYPE=$DB_TYPE
export DB_DSN=$DB_DSN

# ----------------- 2. Database Reset -----------------
echo -e "${YELLOW}🗑️  Resetting database...${NC}"
# Assumes Docker container name 'go-llm-proxy-server-mongo-1' exists from docker-compose
if docker ps | grep -q "go-llm-proxy-server-mongo-1"; then
    docker exec go-llm-proxy-server-mongo-1 mongosh "$DB_DSN" --eval 'db.connections.deleteMany({}); db.provider_models.deleteMany({}); db.virtual_keys.deleteMany({}); db.virtual_key_assignments.deleteMany({});' > /dev/null
    echo -e "${GREEN}✅ Database cleared.${NC}"
else
    echo -e "${RED}❌ MongoDB container not found! Please run 'docker compose up -d' first.${NC}"
    exit 1
fi

# ----------------- 3. Start Server -----------------
echo -e "${YELLOW}⚡ Starting Proxy Server on port $PORT...${NC}"
./llm-proxy serve > proxy_test.log 2>&1 &
PROXY_PID=$!
sleep 3 # Wait for server cold start

if ! kill -0 $PROXY_PID 2>/dev/null; then
    echo -e "${RED}❌ Proxy failed to start. Logs:${NC}"
    cat proxy_test.log
    exit 1
fi

# ----------------- 4. Auto-Seeding & Testing -----------------
echo -e "${YELLOW}🌱 Seeding & Running connectivity tests...${NC}"

# Create the Virtual Key and capture its ID
VK_ID=$(./llm-proxy vkey add --db-type $DB_TYPE --dsn "$DB_DSN" --name "Smoke-Test-Key" --key "$VKEY" | grep -oE "[a-f0-9-]{36}" | tail -1)
echo "Using Virtual Key ID: $VK_ID"
export VK_ID
export VKEY
export PORT
export GREEN RED YELLOW NC

TEST_FAILED=0
echo "---------------------------------------------------"

# Run tests by source-ing provider scripts
# This keeps the environment variables but separates the logic
for script in ./scripts/providers/*.sh; do
    if [ -f "$script" ]; then
        source "$script" || TEST_FAILED=1
    fi
done

echo "---------------------------------------------------"

# ----------------- 6. Cleanup & Exit -----------------
kill $PROXY_PID

if [ $TEST_FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 CONGRATS! All smoke tests passed successfully.${NC}"
    exit 0
else
    echo -e "${RED}💥 Some tests failed. Check logs above.${NC}"
    exit 1
fi

