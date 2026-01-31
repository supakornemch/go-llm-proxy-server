#!/bin/bash

# Load environment if running standalone
if [ -z "$VK_ID" ]; then
    [ -f .env ] && export $(grep -v '^#' .env | xargs)
    DB_TYPE=${DB_TYPE:-"mongodb"}
    DB_DSN=${DB_DSN:-"mongodb://root:examplepassword@localhost:27017/llm_proxy?authSource=admin"}
    PORT=${PORT:-8132}
    # Use timestamp to avoid UNIQUE constraint errors
    TS=$(date +%s)
    VKEY="sk-manual-azure-$TS"
    VK_ID=$(./llm-proxy vkey add --db-type "$DB_TYPE" --dsn "$DB_DSN" --name "Azure-Test-$TS" --key "$VKEY" | grep -oE "[a-f0-9-]{36}" | tail -1)
    GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'
fi

run_test_azure() {
    # Configuration and seeding for Azure
    if [[ -n "$AZURE_FOUNDRY_API_KEY" && -n "$AZURE_FOUNDRY_URL" ]]; then
        echo "Configuring Azure..."
        TS=$(date +%s)
        AZURE_ID=$(./llm-proxy connection add --db-type "$DB_TYPE" --dsn "$DB_DSN" --provider "azure" --name "Azure-Conn-$TS" --endpoint "$AZURE_FOUNDRY_URL" --api-key "$AZURE_FOUNDRY_API_KEY" | grep -oE "[a-f0-9-]{36}" | tail -1)
        M_ID=$(./llm-proxy model add --db-type "$DB_TYPE" --dsn "$DB_DSN" --conn-id "$AZURE_ID" --name "azure-gpt" --remote "gpt-oss-120b" | grep -oE "[a-f0-9-]{36}" | tail -1)
        ./llm-proxy assign --db-type "$DB_TYPE" --dsn "$DB_DSN" --vkey-id "$VK_ID" --model-id "$M_ID" --alias "gpt-oss-120b" --tps 20 > /dev/null

        echo -n "Testing [gpt-oss-120b]... "
        PAYLOAD='{
            "model": "gpt-oss-120b",
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
        echo -e "${YELLOW}⚠️  Skipping Azure (Missing Environment Variables)${NC}"
        return 0
    fi
}

# Distinguish between sourcing and direct execution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    run_test_azure
    exit $?
else
    run_test_azure
fi
