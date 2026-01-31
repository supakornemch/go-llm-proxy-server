#!/bin/bash

# Load environment if running standalone
if [ -z "$VK_ID" ]; then
    [ -f .env ] && export $(grep -v '^#' .env | xargs)
    DB_TYPE=${DB_TYPE:-"mongodb"}
    DB_DSN=${DB_DSN:-"mongodb://root:examplepassword@localhost:27017/llm_proxy?authSource=admin"}
    PORT=${PORT:-8132}
    
    # Use nanoseconds to ensure uniqueness in case of rapid runs
    TS=$(date +%s%N)
    VKEY="sk-manual-openai-$TS"
    echo "Creating seed Virtual Key..."
    VK_OUT=$(./llm-proxy vkey add --db-type "$DB_TYPE" --dsn "$DB_DSN" --name "OpenAI-Test-$TS" --key "$VKEY")
    VK_ID=$(echo "$VK_OUT" | grep -oE "[a-f0-9-]{36}" | tail -1)
    
    if [ -z "$VK_ID" ]; then
        echo -e "\033[0;31m❌ Failed to create Virtual Key. CLI Output:\033[0m"
        echo "$VK_OUT"
        exit 1
    fi
    GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'
fi

run_test_openai() {
    # Configuration and seeding for OpenAI
    if [[ -n "$OPENAI_API_KEY" && -n "$OPENAI_API_ENDPOINT" ]]; then
        echo "Configuring OpenAI connection..."
        TS=$(date +%s%N)
        CONN_OUT=$(./llm-proxy connection add --db-type "$DB_TYPE" --dsn "$DB_DSN" --provider "openai" --name "OpenAI-Conn-$TS" --endpoint "$OPENAI_API_ENDPOINT" --api-key "$OPENAI_API_KEY")
        OPENAI_ID=$(echo "$CONN_OUT" | grep -oE "[a-f0-9-]{36}" | tail -1)
        
        if [ -z "$OPENAI_ID" ]; then
            echo -e "${RED}❌ Failed to create connection. CLI Output:${NC}"
            echo "$CONN_OUT"
            return 1
        fi

        MODEL_OUT=$(./llm-proxy model add --db-type "$DB_TYPE" --dsn "$DB_DSN" --conn-id "$OPENAI_ID" --name "gpt-4.1" --remote "gpt-4.1")
        M_ID=$(echo "$MODEL_OUT" | grep -oE "[a-f0-9-]{36}" | tail -1)
        
        ./llm-proxy assign --db-type "$DB_TYPE" --dsn "$DB_DSN" --vkey-id "$VK_ID" --model-id "$M_ID" --alias "gpt-4.1" --tps 20 > /dev/null

        echo -n "Testing [gpt-4.1]... "
        PAYLOAD='{
            "model": "gpt-4.1",
            "messages": [{"role": "user", "content": "Hello, are you online?"}],
            "temperature": 0.7
        }'

        RESPONSE=$(curl -s -i -X POST "http://localhost:$PORT/v1/chat/completions" \
          -H "Authorization: Bearer $VKEY" \
          -H "Content-Type: application/json" \
          -d "$PAYLOAD")
        
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
        echo -e "${YELLOW}⚠️  Skipping OpenAI (Missing Environment Variables)${NC}"
        return 0
    fi
}

# Distinguish between sourcing and direct execution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    run_test_openai
    exit $?
else
    run_test_openai
fi
