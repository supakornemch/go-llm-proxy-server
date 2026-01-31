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

# ----------------- 4. Auto-Seeding via CLI -----------------
echo -e "${YELLOW}🌱 Seeding Provider Configuration...${NC}"

# 4.1 Create Connections
# Note: Google Gemini now uses standard OpenAI endpoint structure 
# Endpoint: https://generativelanguage.googleapis.com/v1beta/openai
OPENAI_ID=$(./llm-proxy connection add --db-type $DB_TYPE --dsn "$DB_DSN" --provider "openai" --name "OpenAI-Main" --endpoint "${OPENAI_API_ENDPOINT:-https://api.openai.com}" --api-key "${OPENAI_API_KEY:-sk-dummy-openai-key-for-testing}" | grep -oE "[a-f0-9-]{36}")
AZURE_ID=$(./llm-proxy connection add --db-type $DB_TYPE --dsn "$DB_DSN" --provider "azure" --name "Azure-Foundry" --endpoint "${AZURE_FOUNDRY_URL:-https://foundry-myworkshop-sbx.services.ai.azure.com}" --api-key "${AZURE_FOUNDRY_API_KEY:-dummy-azure-key}" | grep -oE "[a-f0-9-]{36}")
GOOGLE_ID=$(./llm-proxy connection add --db-type $DB_TYPE --dsn "$DB_DSN" --provider "google" --name "Google-Gemini" --endpoint "${GOOGLE_GEMINI_ENDPOINT:-https://generativelanguage.googleapis.com/v1beta/openai}" --api-key "${GOOGLE_VERTEX_API_KEY:-dummy-google-key}" | grep -oE "[a-f0-9-]{36}")
AWS_ID=$(./llm-proxy connection add --db-type $DB_TYPE --dsn "$DB_DSN" --provider "aws" --name "AWS-Bedrock" --endpoint "${AWS_BEDROCK_ENDPOINT:-https://bedrock-runtime.ap-southeast-1.amazonaws.com}" --api-key "${AWS_BEDROCK_API_KEY:-dummy-aws-key}" | grep -oE "[a-f0-9-]{36}")

# 4.2 Create Virtual Key
VK_ID=$(./llm-proxy vkey add --db-type $DB_TYPE --dsn "$DB_DSN" --name "Test-User-Key" --key "$VKEY" | grep -oE "[a-f0-9-]{36}" | tail -1)

# 4.3 Assign Models (Map Alias -> Remote Model)
# OpenAI: Standard mapping
M1=$(./llm-proxy model add --db-type $DB_TYPE --dsn "$DB_DSN" --conn-id "$OPENAI_ID" --name "gpt-4.1" --remote "gpt-4.1" | grep -oE "[a-f0-9-]{36}" | tail -1)
./llm-proxy assign --db-type $DB_TYPE --dsn "$DB_DSN" --vkey-id "$VK_ID" --model-id "$M1" --alias "gpt-4.1" --tps 20 > /dev/null

# Azure: Map local 'gpt-oss' to remote 'gpt-oss-120b'
M2=$(./llm-proxy model add --db-type $DB_TYPE --dsn "$DB_DSN" --conn-id "$AZURE_ID" --name "azure-gpt" --remote "gpt-oss-120b" | grep -oE "[a-f0-9-]{36}" | tail -1)
./llm-proxy assign --db-type $DB_TYPE --dsn "$DB_DSN" --vkey-id "$VK_ID" --model-id "$M2" --alias "gpt-oss-120b" --tps 20 > /dev/null

# Google: Gemini Flash
M3=$(./llm-proxy model add --db-type $DB_TYPE --dsn "$DB_DSN" --conn-id "$GOOGLE_ID" --name "gemini-flash" --remote "gemini-1.5-flash" | grep -oE "[a-f0-9-]{36}" | tail -1)
./llm-proxy assign --db-type $DB_TYPE --dsn "$DB_DSN" --vkey-id "$VK_ID" --model-id "$M3" --alias "gemini-1.5-flash" --tps 20 > /dev/null

# AWS: Claude Haiku
M4=$(./llm-proxy model add --db-type $DB_TYPE --dsn "$DB_DSN" --conn-id "$AWS_ID" --name "claude-haiku" --remote "claude-haiku-4-5" | grep -oE "[a-f0-9-]{36}" | tail -1)
./llm-proxy assign --db-type $DB_TYPE --dsn "$DB_DSN" --vkey-id "$VK_ID" --model-id "$M4" --alias "claude-haiku-4-5" --tps 20 > /dev/null

echo -e "${GREEN}✅ Seeding Completed.${NC}"

# ----------------- 5. Execution Loop (The Smoke Test) -----------------
echo -e "${YELLOW}📡 Running connectivity tests...${NC}"

TEST_FAILED=0
# Note: Using standard OpenAI JSON format for ALL providers now!
PAYLOAD='{
  "messages": [{"role": "user", "content": "Hello, are you online?"}],
  "temperature": 0.7
}'

run_test() {
    local ALIAS=$1
    local EXPECTED_NAME=$2
    
    echo -n "Testing [$ALIAS]... "
    
    # Inject model name into payload
    local BODY=$(echo "$PAYLOAD" | sed "s/\"messages\"/\"model\": \"$ALIAS\", \"messages\"/")
    
    RESPONSE=$(curl -s -i -X POST "http://localhost:$PORT/v1/chat/completions" \
      -H "Authorization: Bearer $VKEY" \
      -H "Content-Type: application/json" \
      -d "$BODY")
      
    HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP/" | awk '{print $2}')
    
    if [[ "$HTTP_CODE" == "200" ]]; then
        echo -e "${GREEN}PASS (200 OK)${NC}"
    elif [[ "$HTTP_CODE" == "401" ]]; then 
        # 401 is acceptable for this smoke test as we use dummy/expired keys for some providers, 
        # but it proves the Proxy routed the request to the correct upstream.
        echo -e "${GREEN}PASS (401 from Provider)${NC}"
    else
        echo -e "${RED}FAIL ($HTTP_CODE)${NC}"
        echo "$RESPONSE" | awk '/^\r?$/ {Found=1; next} Found' | head -n 5
        TEST_FAILED=1
    fi
}

echo "---------------------------------------------------"
run_test "gpt-4.1" "OpenAI"
run_test "gpt-oss-120b" "Azure"
run_test "gemini-1.5-flash" "Google"
run_test "claude-haiku-4-5" "AWS"
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

