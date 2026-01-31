#!/bin/bash

# Load environment if running standalone
if [ -z "$VK_ID" ]; then
    [ -f .env ] && export $(grep -v '^#' .env | xargs)
    DB_TYPE=${DB_TYPE:-"mongodb"}
    DB_DSN=${DB_DSN:-"mongodb://root:examplepassword@localhost:27017/llm_proxy?authSource=admin"}
    PORT=${PORT:-8132}
    
    # Use nanoseconds to ensure uniqueness in case of rapid runs
    TS=$(date +%s%N)
    VKEY="sk-manual-google-$TS"
    echo "Creating seed Virtual Key..."
    VK_OUT=$(./llm-proxy vkey add --db-type "$DB_TYPE" --dsn "$DB_DSN" --name "Google-Test-$TS" --key "$VKEY")
    VK_ID=$(echo "$VK_OUT" | grep -oE "[a-f0-9-]{36}" | tail -1)
    
    if [ -z "$VK_ID" ]; then
        echo -e "\033[0;31m❌ Failed to create Virtual Key. CLI Output:\033[0m"
        echo "$VK_OUT"
        exit 1
    fi
    GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'
fi

run_test_google() {
    # Configuration and seeding for Google (Gemini)
    if [[ -n "$GOOGLE_VERTEX_API_KEY" && -n "$GOOGLE_GEMINI_ENDPOINT" ]]; then
        echo "Configuring Google connection..."
        TS=$(date +%s%N)
        CONN_OUT=$(./llm-proxy connection add --db-type "$DB_TYPE" --dsn "$DB_DSN" --provider "google" --name "Google-Conn-$TS" --endpoint "$GOOGLE_GEMINI_ENDPOINT" --api-key "$GOOGLE_VERTEX_API_KEY")
        GOOGLE_ID=$(echo "$CONN_OUT" | grep -oE "[a-f0-9-]{36}" | tail -1)

        if [ -z "$GOOGLE_ID" ]; then
            echo -e "${RED}❌ Failed to create connection. CLI Output:${NC}"
            echo "$CONN_OUT"
            return 1
        fi

        MODEL_OUT=$(./llm-proxy model add --db-type "$DB_TYPE" --dsn "$DB_DSN" --conn-id "$GOOGLE_ID" --name "gemini-3" --remote "gemini-3-flash-preview")
        M_ID=$(echo "$MODEL_OUT" | grep -oE "[a-f0-9-]{36}" | tail -1)
        
        ./llm-proxy assign --db-type "$DB_TYPE" --dsn "$DB_DSN" --vkey-id "$VK_ID" --model-id "$M_ID" --alias "gemini-3-flash-preview" --tps 20 > /dev/null

        echo -n "Testing [gemini-3-flash-preview]... "
        TARGET_PATH="/v1/publishers/google/models/gemini-3-flash-preview"
        # Using the native Gemini body as requested in previous step
        BODY='{
            "contents": [
                {
                    "role": "user",
                    "parts": [
                        {
                            "text": "hihi"
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

        RESPONSE=$(curl -s -i -X POST "http://localhost:$PORT$TARGET_PATH?key=PLACEHOLDER_KEY" \
          -H "Authorization: Bearer $VKEY" \
          -H "Content-Type: application/json" \
          -d "$BODY")
        
        HTTP_CODE=$(echo "$RESPONSE" | grep "HTTP/" | awk '{print $2}')
        
        if [[ "$HTTP_CODE" == "200" ]]; then
            echo -e "${GREEN}PASS (200 OK)${NC}"
            return 0
        else
            echo -e "${RED}FAIL ($HTTP_CODE)${NC}"
            echo -e "${YELLOW}Response from Provider:${NC}"
            echo "$RESPONSE" | awk '/^\r?$/ {Found=1; next} Found'
            return 1
        fi
    else
        echo -e "${YELLOW}⚠️  Skipping Google (Missing Environment Variables)${NC}"
        return 0
    fi
}

# Distinguish between sourcing and direct execution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    run_test_google
    exit $?
else
    run_test_google
fi
