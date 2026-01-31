#!/bin/bash

# =================================================================================================
# LLM Proxy - Quick Start with Auto Setup & Python Execution
# =================================================================================================
# Automated setup: Create Connections, Models, Virtual Keys, and run Python examples immediately
# Usage: ./quickstart.sh --provider openai|google|azure|all
# =================================================================================================

set -e

# ============ Configuration ============
PORT=8132
DB_TYPE=${DB_TYPE:-mongodb}
DB_DSN=${DB_DSN:-"mongodb://root:examplepassword@localhost:27017/llm_proxy?authSource=admin"}
VKEY="vk-quickstart-$(date +%s)"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# ============ Functions ============
log_step() {
    echo -e "${BLUE}▶${NC} $1"
}

log_success() {
    echo -e "${GREEN}✅${NC} $1"
}

log_error() {
    echo -e "${RED}❌${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠️${NC} $1"
}

check_proxy_running() {
    if ! curl -s http://localhost:$PORT/health > /dev/null 2>&1; then
        log_error "Proxy Server is not running on port $PORT"
        log_step "Start it with: ./llm-proxy serve"
        exit 1
    fi
    log_success "Proxy Server is running"
}

create_connection() {
    local PROVIDER=$1
    local NAME=$2
    local ENDPOINT=$3
    local API_KEY=$4
    
    if [ -z "$API_KEY" ]; then
        log_warning "Skipping $NAME (API Key not set)"
        return
    fi
    
    log_step "Creating connection: $NAME"
    local CONN_ID=$(./llm-proxy connection add \
        --db-type "$DB_TYPE" \
        --dsn "$DB_DSN" \
        --provider "$PROVIDER" \
        --name "$NAME" \
        --endpoint "$ENDPOINT" \
        --api-key "$API_KEY" 2>/dev/null | grep -oE "[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}" | head -1)
    
    if [ -z "$CONN_ID" ]; then
        log_error "Failed to create connection $NAME"
        return
    fi
    
    log_success "Connection created: $CONN_ID"
    echo "$CONN_ID"
}

create_model() {
    local CONN_ID=$1
    local MODEL_NAME=$2
    local REMOTE_NAME=$3
    
    if [ -z "$CONN_ID" ]; then
        return
    fi
    
    log_step "Adding model: $MODEL_NAME → $REMOTE_NAME"
    local MODEL_ID=$(./llm-proxy model add \
        --db-type "$DB_TYPE" \
        --dsn "$DB_DSN" \
        --conn-id "$CONN_ID" \
        --name "$MODEL_NAME" \
        --remote "$REMOTE_NAME" 2>/dev/null | grep -oE "[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}" | head -1)
    
    if [ -z "$MODEL_ID" ]; then
        log_error "Failed to create model $MODEL_NAME"
        return
    fi
    
    log_success "Model created: $MODEL_ID"
    echo "$MODEL_ID"
}

assign_model() {
    local VKEY_ID=$1
    local MODEL_ID=$2
    local ALIAS=$3
    local TPS=${4:-50}
    
    if [ -z "$MODEL_ID" ]; then
        return
    fi
    
    log_step "Assigning model: $ALIAS (TPS: $TPS)"
    ./llm-proxy assign \
        --db-type "$DB_TYPE" \
        --dsn "$DB_DSN" \
        --vkey-id "$VKEY_ID" \
        --model-id "$MODEL_ID" \
        --alias "$ALIAS" \
        --tps "$TPS" > /dev/null 2>&1
    
    log_success "Model assigned: $ALIAS"
}

run_example() {
    local EXAMPLE=$1
    local VKEY=$2
    local ALIAS=$3
    
    if [ ! -f "examples/$EXAMPLE" ]; then
        log_warning "Example file not found: $EXAMPLE"
        return
    fi
    
    log_step "Running example: $EXAMPLE (VKey: $VKEY, Model: $ALIAS)"
    echo ""
    
    # Inject Virtual Key and Model Alias into the example
    python3 -c "
import sys
import os

# Read the example file
with open('examples/$EXAMPLE', 'r') as f:
    code = f.read()

# Replace placeholders
code = code.replace('vk-frontend-app', '$VKEY')
code = code.replace('vk-google-app', '$VKEY')
code = code.replace('vk-azure-app', '$VKEY')
code = code.replace('vk-my-app', '$VKEY')
code = code.replace('gpt-4-turbo', '$ALIAS')
code = code.replace('gpt-4o', '$ALIAS')
code = code.replace('gemini-3-flash', '$ALIAS')
code = code.replace('gemini-2-flash', '$ALIAS')

# Execute
exec(code)
" 2>&1 || true
    
    echo ""
}

setup_openai() {
    log_step "🔴 Setting up OpenAI"
    
    local CONN_ID=$(create_connection "openai" "OpenAI-Main" \
        "${OPENAI_API_ENDPOINT:-https://api.openai.com}" \
        "$OPENAI_API_KEY")
    
    if [ -z "$CONN_ID" ]; then
        return
    fi
    
    local MODEL_ID=$(create_model "$CONN_ID" "gpt-4-turbo" "gpt-4-turbo-preview")
    assign_model "$VKEY_ID" "$MODEL_ID" "gpt-4-turbo"
    
    OPENAI_MODEL_ALIAS="gpt-4-turbo"
}

setup_azure() {
    log_step "🔵 Setting up Azure OpenAI"
    
    local CONN_ID=$(create_connection "azure" "Azure-Main" \
        "$AZURE_OPENAI_ENDPOINT" \
        "$AZURE_OPENAI_API_KEY")
    
    if [ -z "$CONN_ID" ]; then
        return
    fi
    
    local MODEL_ID=$(create_model "$CONN_ID" "gpt-4o" "gpt-4o")
    assign_model "$VKEY_ID" "$MODEL_ID" "gpt-4o"
    
    AZURE_MODEL_ALIAS="gpt-4o"
}

setup_google() {
    log_step "🟢 Setting up Google Vertex AI"
    
    local CONN_ID=$(create_connection "google" "Google-Vertex" \
        "${GOOGLE_GEMINI_ENDPOINT:-https://aiplatform.googleapis.com}" \
        "$GOOGLE_VERTEX_API_KEY")
    
    if [ -z "$CONN_ID" ]; then
        return
    fi
    
    local MODEL_ID=$(create_model "$CONN_ID" "gemini-3-flash" "gemini-3-flash-preview")
    assign_model "$VKEY_ID" "$MODEL_ID" "gemini-3-flash-preview"
    
    GOOGLE_MODEL_ALIAS="gemini-3-flash-preview"
}

# ============ Main Script ============
main() {
    local PROVIDER_CHOICE=${1:-all}
    
    echo ""
    echo -e "${YELLOW}================================${NC}"
    echo -e "${YELLOW}🚀 LLM Proxy Quick Start${NC}"
    echo -e "${YELLOW}================================${NC}"
    echo ""
    
    # 1. Check if Proxy is running
    log_step "Checking Proxy Server..."
    check_proxy_running
    echo ""
    
    # 2. Create Virtual Key
    log_step "Creating Virtual Key..."
    VKEY_ID=$(./llm-proxy vkey add \
        --db-type "$DB_TYPE" \
        --dsn "$DB_DSN" \
        --name "QuickStart-$(date +%s)" \
        --key "$VKEY" 2>/dev/null | grep -oE "[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}" | head -1)
    
    if [ -z "$VKEY_ID" ]; then
        log_error "Failed to create Virtual Key"
        exit 1
    fi
    
    log_success "Virtual Key created: $VKEY_ID"
    echo "  Virtual Key: $VKEY"
    echo ""
    
    # 3. Setup Providers
    case $PROVIDER_CHOICE in
        openai)
            setup_openai
            ;;
        azure)
            setup_azure
            ;;
        google)
            setup_google
            ;;
        all)
            setup_openai
            echo ""
            setup_azure
            echo ""
            setup_google
            echo ""
            ;;
        *)
            log_error "Unknown provider: $PROVIDER_CHOICE"
            echo "Usage: ./quickstart.sh [openai|azure|google|all]"
            exit 1
            ;;
    esac
    
    echo ""
    echo -e "${GREEN}================================${NC}"
    echo -e "${GREEN}✨ Setup Complete!${NC}"
    echo -e "${GREEN}================================${NC}"
    echo ""
    echo "Virtual Key: $VKEY"
    echo "Base URL: http://localhost:$PORT"
    echo ""
    echo "Ready to execute examples! 🎯"
    echo ""
    
    # 4. Ask to run examples
    read -p "Run Python examples now? (y/n) " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo ""
        
        if [ ! -z "$OPENAI_MODEL_ALIAS" ]; then
            echo -e "${BLUE}📌 Running OpenAI Example${NC}"
            run_example "example_openai.py" "$VKEY" "$OPENAI_MODEL_ALIAS"
        fi
        
        if [ ! -z "$AZURE_MODEL_ALIAS" ]; then
            echo -e "${BLUE}📌 Running Azure Example${NC}"
            run_example "example_azure.py" "$VKEY" "$AZURE_MODEL_ALIAS"
        fi
        
        if [ ! -z "$GOOGLE_MODEL_ALIAS" ]; then
            echo -e "${BLUE}📌 Running Google Vertex Example${NC}"
            run_example "example_google_vertex_http.py" "$VKEY" "$GOOGLE_MODEL_ALIAS"
        fi
        
        echo -e "${GREEN}🎉 All examples executed!${NC}"
    fi
}

main "$@"
