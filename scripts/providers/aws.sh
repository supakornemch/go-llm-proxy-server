#!/bin/bash

# Configuration and seeding for AWS (Bedrock)
if [[ -n "$AWS_BEDROCK_API_KEY" && -n "$AWS_BEDROCK_ENDPOINT" ]]; then
    echo "Configuring AWS..."
    AWS_ID=$(./llm-proxy connection add --db-type "$DB_TYPE" --dsn "$DB_DSN" --provider "aws" --name "AWS-Bedrock" --endpoint "$AWS_BEDROCK_ENDPOINT" --api-key "$AWS_BEDROCK_API_KEY" | grep -oE "[a-f0-9-]{36}")
    M_ID=$(./llm-proxy model add --db-type "$DB_TYPE" --dsn "$DB_DSN" --conn-id "$AWS_ID" --name "claude-haiku" --remote "claude-haiku-4-5" | grep -oE "[a-f0-9-]{36}" | tail -1)
    ./llm-proxy assign --db-type "$DB_TYPE" --dsn "$DB_DSN" --vkey-id "$VK_ID" --model-id "$M_ID" --alias "claude-haiku-4-5" --tps 20 > /dev/null

    echo -n "Testing [claude-haiku-4-5]... "
    PAYLOAD='{
        "model": "claude-haiku-4-5",
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
    else
        echo -e "${RED}FAIL ($HTTP_CODE)${NC}"
        echo -e "${YELLOW}Response from Provider:${NC}"
        echo "$RESPONSE" | awk '/^\r?$/ {Found=1; next} Found'
        return 1
    fi
    return 0
else
    echo -e "${YELLOW}⚠️  Skipping AWS (Missing Environment Variables)${NC}"
fi
