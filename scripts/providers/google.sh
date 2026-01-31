#!/bin/bash

# Configuration and seeding for Google (Gemini)
if [[ -n "$GOOGLE_VERTEX_API_KEY" && -n "$GOOGLE_GEMINI_ENDPOINT" ]]; then
    echo "Configuring Gemini 3..."
    GOOGLE_ID=$(./llm-proxy connection add --db-type "$DB_TYPE" --dsn "$DB_DSN" --provider "google" --name "Google-Gemini" --endpoint "$GOOGLE_GEMINI_ENDPOINT" --api-key "$GOOGLE_VERTEX_API_KEY" | grep -oE "[a-f0-9-]{36}")
    M_ID=$(./llm-proxy model add --db-type "$DB_TYPE" --dsn "$DB_DSN" --conn-id "$GOOGLE_ID" --name "gemini-3" --remote "gemini-3-flash-preview" | grep -oE "[a-f0-9-]{36}" | tail -1)
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
    else
        echo -e "${RED}FAIL ($HTTP_CODE)${NC}"
        echo -e "${YELLOW}Response from Provider:${NC}"
        echo "$RESPONSE" | awk '/^\r?$/ {Found=1; next} Found'
        return 1
    fi
    return 0
else
    echo -e "${YELLOW}⚠️  Skipping Google (Missing Environment Variables)${NC}"
fi
