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
DB_TYPE=sqlite
DB_DSN="llm_proxy.db"
PROVIDER_FILTER="all"

# Parse arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --provider) PROVIDER_FILTER="$2"; shift ;;
        *) echo "Unknown parameter passed: $1"; exit 1 ;;
    esac
    shift
done

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
if [ -f "$DB_DSN" ]; then
    rm "$DB_DSN"
    echo "Deleted existing database file: $DB_DSN"
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

# Create the Virtual Key and capture its ID
VK_ID=$(./llm-proxy vkey add --db-type $DB_TYPE --dsn "$DB_DSN" --name "Smoke-Test-Key" --key "$VKEY" | grep -oE "[a-f0-9-]{36}" | tail -1)
echo "Using Virtual Key ID: $VK_ID"

# 4.1 Create Connections & Models (Conditional)
MODELS_TO_TEST=()

# 4.1.1 OpenAI
if [[ "$PROVIDER_FILTER" == "all" || "$PROVIDER_FILTER" == "openai" ]]; then
    if [[ -n "$OPENAI_API_KEY" && -n "$OPENAI_API_ENDPOINT" ]]; then
        echo "Configuring OpenAI..."
        OPENAI_ID=$(./llm-proxy connection add --db-type $DB_TYPE --dsn "$DB_DSN" --provider "openai" --name "OpenAI-Main" --endpoint "$OPENAI_API_ENDPOINT" --api-key "$OPENAI_API_KEY" | grep -oE "[a-f0-9-]{36}")
        M_ID=$(./llm-proxy model add --db-type $DB_TYPE --dsn "$DB_DSN" --conn-id "$OPENAI_ID" --name "gpt-4.1" --remote "gpt-4.1" | grep -oE "[a-f0-9-]{36}" | tail -1)
        ./llm-proxy assign --db-type $DB_TYPE --dsn "$DB_DSN" --vkey-id "$VK_ID" --model-id "$M_ID" --alias "gpt-4.1" --tps 20 > /dev/null
        MODELS_TO_TEST+=("gpt-4.1")
    else
        echo -e "${YELLOW}⚠️  Skipping OpenAI (Missing Environment Variables)${NC}"
    fi
fi

# 4.1.2 Azure
if [[ "$PROVIDER_FILTER" == "all" || "$PROVIDER_FILTER" == "azure" ]]; then
    if [[ -n "$AZURE_FOUNDRY_API_KEY" && -n "$AZURE_FOUNDRY_URL" ]]; then
        echo "Configuring Azure..."
        AZURE_ID=$(./llm-proxy connection add --db-type $DB_TYPE --dsn "$DB_DSN" --provider "azure" --name "Azure-Foundry" --endpoint "$AZURE_FOUNDRY_URL" --api-key "$AZURE_FOUNDRY_API_KEY" | grep -oE "[a-f0-9-]{36}")
        M_ID=$(./llm-proxy model add --db-type $DB_TYPE --dsn "$DB_DSN" --conn-id "$AZURE_ID" --name "azure-gpt" --remote "gpt-oss-120b" | grep -oE "[a-f0-9-]{36}" | tail -1)
        ./llm-proxy assign --db-type $DB_TYPE --dsn "$DB_DSN" --vkey-id "$VK_ID" --model-id "$M_ID" --alias "gpt-oss-120b" --tps 20 > /dev/null
        MODELS_TO_TEST+=("gpt-oss-120b")
    else
        echo -e "${YELLOW}⚠️  Skipping Azure (Missing Environment Variables)${NC}"
    fi
fi

# 4.1.3 Google (GEMINI 3)
if [[ "$PROVIDER_FILTER" == "all" || "$PROVIDER_FILTER" == "google" ]]; then
    if [[ -n "$GOOGLE_VERTEX_API_KEY" && -n "$GOOGLE_GEMINI_ENDPOINT" ]]; then
        echo "Configuring Gemini 3..."
        GOOGLE_ID=$(./llm-proxy connection add --db-type $DB_TYPE --dsn "$DB_DSN" --provider "google" --name "Google-Gemini" --endpoint "$GOOGLE_GEMINI_ENDPOINT" --api-key "$GOOGLE_VERTEX_API_KEY" | grep -oE "[a-f0-9-]{36}")
        M_ID=$(./llm-proxy model add --db-type $DB_TYPE --dsn "$DB_DSN" --conn-id "$GOOGLE_ID" --name "gemini-3" --remote "gemini-3-flash-preview" | grep -oE "[a-f0-9-]{36}" | tail -1)
        ./llm-proxy assign --db-type $DB_TYPE --dsn "$DB_DSN" --vkey-id "$VK_ID" --model-id "$M_ID" --alias "gemini-3-flash-preview" --tps 20 > /dev/null
        MODELS_TO_TEST+=("gemini-3-flash-preview")
    else
        echo -e "${YELLOW}⚠️  Skipping Google (Missing Environment Variables)${NC}"
    fi
fi

# 4.1.4 AWS
if [[ "$PROVIDER_FILTER" == "all" || "$PROVIDER_FILTER" == "aws" ]]; then
    if [[ -n "$AWS_BEDROCK_API_KEY" && -n "$AWS_BEDROCK_ENDPOINT" ]]; then
        echo "Configuring AWS..."
        AWS_ID=$(./llm-proxy connection add --db-type $DB_TYPE --dsn "$DB_DSN" --provider "aws" --name "AWS-Bedrock" --endpoint "$AWS_BEDROCK_ENDPOINT" --api-key "$AWS_BEDROCK_API_KEY" | grep -oE "[a-f0-9-]{36}")
        M_ID=$(./llm-proxy model add --db-type $DB_TYPE --dsn "$DB_DSN" --conn-id "$AWS_ID" --name "claude-haiku" --remote "claude-haiku-4-5" | grep -oE "[a-f0-9-]{36}" | tail -1)
        ./llm-proxy assign --db-type $DB_TYPE --dsn "$DB_DSN" --vkey-id "$VK_ID" --model-id "$M_ID" --alias "claude-haiku-4-5" --tps 20 > /dev/null
        MODELS_TO_TEST+=("claude-haiku-4-5")
    else
        echo -e "${YELLOW}⚠️  Skipping AWS (Missing Environment Variables)${NC}"
    fi
fi

echo -e "${GREEN}✅ Seeding Completed.${NC}"

# ----------------- 5. Execution Loop (The Smoke Test) -----------------
echo -e "${YELLOW}📡 Running connectivity tests...${NC}"

TEST_FAILED=0
# Note: Using standard OpenAI JSON format for ALL providers now!
PAYLOAD='{
  "messages": [{"role": "user", "content": "Please say hello to me!"}],
  "temperature": 0.7
}'

run_test() {
    local ALIAS=$1
    echo -n "Testing [$ALIAS]... "
    
    local BODY
    local TARGET_PATH="/v1/chat/completions"
    
    if [[ "$ALIAS" == *"gemini-3"* ]]; then
        # Use exact payload from user's Gemini 3 official curl sample
        # Note: The path is the Vertex AI style path requested by user
        TARGET_PATH="/v1/publishers/google/models/$ALIAS:streamGenerateContent"
        BODY='{
            "contents": [
                {
                    "role": "user",
                    "parts": [
                        {
                            "text": "Please say hello to me!"
                        }
                    ]
                }
            ],
            "generationConfig": {
                "temperature": 1,
                "maxOutputTokens": 65535,
                "topP": 0.95,
                "thinkingConfig": {
                    "thinkingLevel": "LOW"
                }
            }
        }'
    else
        # OpenAI style
        BODY=$(echo "$PAYLOAD" | sed "s/\"messages\"/\"model\": \"$ALIAS\", \"messages\"/")
    fi
    
    # We also include ?key=... in the URL to the proxy, just like the user's sample
    # Our proxy should handle stripping/replacing it with the backend key.
    RESPONSE=$(curl -s -i -X POST "http://localhost:$PORT$TARGET_PATH?key=PLACEHOLDER_KEY" \
      -H "Authorization: Bearer $VKEY" \
      -H "Content-Type: application/json" \
      -d "$BODY")
      
    HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP/" | awk '{print $2}')
    BODY_DATA=$(echo "$RESPONSE" | awk '/^\r?$/ {Found=1; next} Found')
    
    # Print basic status
    if [[ "$HTTP_CODE" == "200" ]]; then
        echo -e "${GREEN}SUCCESS ($HTTP_CODE)${NC}"
    else
        echo -e "${RED}PROVIDER_RESPONSE ($HTTP_CODE)${NC}"
    fi

    echo -e "${YELLOW}--- RAW BODY START ---${NC}"
    if echo "$BODY_DATA" | jq . &>/dev/null; then
        echo "$BODY_DATA" | jq .
    else
        echo "$BODY_DATA"
    fi
    echo -e "${YELLOW}--- RAW BODY END ---${NC}"
    echo "---------------------------------------------------"
    
    # We don't mark as FAILED for 4xx/5xx from provider because user might want to see them
    # But for proxy connection errors (no HTTP_CODE), it's a real failure
    if [[ -z "$HTTP_CODE" ]]; then
        TEST_FAILED=1
    fi
}

echo "---------------------------------------------------"
if [ ${#MODELS_TO_TEST[@]} -eq 0 ]; then
    echo -e "${RED}❌ No models were configured. Testing skipped.${NC}"
    TEST_FAILED=1
else
    for MODEL in "${MODELS_TO_TEST[@]}"; do
        run_test "$MODEL"
    done
fi
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

