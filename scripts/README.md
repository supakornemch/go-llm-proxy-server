# Quick Start Scripts

Automated setup and execution scripts for LLM Proxy

## 📚 Scripts

### 1. `quickstart.sh` - Bash Version
Shell script that automates the complete setup process

```bash
# Setup all providers (OpenAI, Azure, Google)
./scripts/quickstart.sh all

# Setup specific provider
./scripts/quickstart.sh openai
./scripts/quickstart.sh azure
./scripts/quickstart.sh google
```

**Features:**
- Creates Connections, Models, Virtual Keys via CLI
- Automatically runs Python examples
- Color-coded output
- Interactive prompt to execute examples

### 2. `quickstart.py` - Python Version
Python script with subprocess integration

```bash
# Setup all providers
python3 scripts/quickstart.py --provider all

# Setup specific provider
python3 scripts/quickstart.py --provider openai
python3 scripts/quickstart.py --provider azure
python3 scripts/quickstart.py --provider google

# Skip running examples
python3 scripts/quickstart.py --provider all --skip-examples
```

**Features:**
- Pure Python implementation
- CLI command execution via subprocess
- Auto-inject Virtual Key and Model Alias into examples
- Better error handling

## 🚀 Prerequisites

Before running quickstart:

1. **Start Proxy Server:**
   ```bash
   ./llm-proxy serve
   ```

2. **Set Environment Variables:**
   ```bash
   # For OpenAI
   export OPENAI_API_KEY="sk-proj-..."
   export OPENAI_API_ENDPOINT="https://api.openai.com"

   # For Azure
   export AZURE_OPENAI_API_KEY="your-key"
   export AZURE_OPENAI_ENDPOINT="https://xxx.openai.azure.com"

   # For Google
   export GOOGLE_VERTEX_API_KEY="AQ...."
   export GOOGLE_GEMINI_ENDPOINT="https://aiplatform.googleapis.com"
   ```

3. **Install Dependencies:**
   ```bash
   python3 examples/setup.py
   ```

## 🎯 What Happens

When you run quickstart:

1. ✅ Checks if Proxy Server is running
2. ✅ Creates Virtual Key (with unique ID)
3. ✅ Creates Connection for each enabled provider
4. ✅ Creates Models and assigns them to Virtual Key
5. ✅ Runs Python examples with auto-injected credentials
6. ✅ Displays results

## 📊 Output Example

```
================================
🚀 LLM Proxy Quick Start
================================

▶ Checking Proxy Server...
✅ Proxy Server is running

▶ Creating Virtual Key...
✅ Virtual Key created: 550e8400-e29b-41d4-a716-446655440000
  Key: vk-quickstart-1707154200

▶ 🔴 Setting up OpenAI
▶ Creating connection: OpenAI-Main
✅ Connection created: 550e8400-e29b-41d4-a716-446655440001
▶ Adding model: gpt-4-turbo → gpt-4-turbo-preview
✅ Model created: 550e8400-e29b-41d4-a716-446655440002
▶ Assigning model: gpt-4-turbo (TPS: 50)
✅ Model assigned: gpt-4-turbo

...

================================
✨ Setup Complete!
================================

Virtual Key: vk-quickstart-1707154200
Base URL: http://localhost:8132

Run Python examples now? (y/n): y

▶ Running example: example_openai.py
  Model: gpt-4-turbo, Key: vk-quickstart-1707154200

🤖 OpenAI via Proxy - Chat Completion Example

✅ Model: gpt-4-turbo-preview
💬 Response: Hello! I'm here and ready to help. What can I assist you with today?
📊 Usage: 8 prompt tokens, 18 completion tokens

🎉 All examples executed!
```

## 💡 Tips

- **Environment Variables:** Both scripts read from `$OPENAI_API_KEY`, `$AZURE_OPENAI_API_KEY`, `$GOOGLE_VERTEX_API_KEY`
- **Database Configuration:** Customizable via `DB_TYPE` and `DB_DSN` env vars
- **Port:** Default is 8132, change via `PORT` env var
- **Skip Examples:** Run with `--skip-examples` to only setup, not execute

## 📚 More Information

- See [examples/README.md](../examples/README.md) for individual example details
- See [docs/GUIDE_TH.md](../docs/GUIDE_TH.md) for detailed setup guide
