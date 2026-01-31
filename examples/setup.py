#!/usr/bin/env python3
"""
Setup Guide - ติดตั้ง Dependencies และสร้าง Virtual Key
รันสคริปต์นี้ก่อนใช้ตัวอย่างอื่น ๆ
"""

import subprocess
import sys

def run_command(cmd):
    """รัน shell command"""
    try:
        result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
        return result.returncode == 0, result.stdout, result.stderr
    except Exception as e:
        return False, "", str(e)


def install_dependencies():
    """ติดตั้ง Python dependencies"""
    print("📦 Installing Python dependencies...\n")

    packages = [
        ("openai", "OpenAI SDK"),
        ("azure-openai", "Azure OpenAI SDK"),
        ("google-generativeai", "Google Generative AI SDK"),
        ("boto3", "AWS Bedrock SDK"),
        ("requests", "HTTP Requests library"),
    ]

    for package, name in packages:
        print(f"  Installing {name} ({package})...")
        success, stdout, stderr = run_command(f"{sys.executable} -m pip install {package} -q")
        if success:
            print(f"  ✅ {name} installed successfully")
        else:
            print(f"  ⚠️  Failed to install {name}")
            print(f"     Error: {stderr}")

    print()


def create_virtual_key():
    """สร้าง Virtual Key สำหรับ Client"""
    print("🔑 Creating Virtual Key...\n")

    print("  Example commands (run on Proxy Server):\n")

    examples = [
        ('vkey add --name "Frontend-App" --key "vk-frontend-app"', "Frontend App"),
        ('vkey add --name "Google-App" --key "vk-google-app"', "Google Vertex/Gemini"),
        ('vkey add --name "Azure-App" --key "vk-azure-app"', "Azure OpenAI"),
        ('vkey add --name "MyApp" --key "vk-my-app"', "General Purpose"),
    ]

    for cmd, desc in examples:
        print(f"  # {desc}")
        print(f"  ./llm-proxy {cmd}\n")


def show_setup_instructions():
    """แสดง instructions สำหรับการตั้งค่า"""
    print("\n" + "=" * 60)
    print("📋 Setup Instructions")
    print("=" * 60 + "\n")

    instructions = """
1️⃣  START PROXY SERVER:
    ./llm-proxy serve

2️⃣  ADD CONNECTIONS (in another terminal):
    # OpenAI
    ./llm-proxy connection add \\
      --name "OpenAI" \\
      --provider "openai" \\
      --endpoint "https://api.openai.com" \\
      --api-key "sk-proj-..."

    # Azure
    ./llm-proxy connection add \\
      --name "Azure" \\
      --provider "azure" \\
      --endpoint "https://xxx.openai.azure.com" \\
      --api-key "your-azure-key"

    # Google Vertex
    ./llm-proxy connection add \\
      --name "Vertex" \\
      --provider "google" \\
      --endpoint "https://aiplatform.googleapis.com" \\
      --api-key "AQ...."

3️⃣  ADD MODELS:
    ./llm-proxy model add \\
      --conn-id "conn-xxx" \\
      --name "gpt-4-turbo" \\
      --remote "gpt-4-turbo-preview"

4️⃣  CREATE VIRTUAL KEYS:
    ./llm-proxy vkey add --name "App1" --key "vk-app1"

5️⃣  ASSIGN MODELS TO KEYS:
    ./llm-proxy assign \\
      --vkey-id "vkey-xxx" \\
      --model-id "model-xxx" \\
      --alias "gpt-4-turbo" \\
      --tps 50

6️⃣  RUN EXAMPLES:
    python3 examples/example_openai.py
    python3 examples/example_google_vertex_http.py
    python3 examples/example_raw_http.py

"""

    print(instructions)


def main():
    print("\n" + "=" * 60)
    print("🚀 LLM Proxy - Setup & Installation")
    print("=" * 60 + "\n")

    # 1. Install dependencies
    install_dependencies()

    # 2. Show Virtual Key creation examples
    create_virtual_key()

    # 3. Show setup instructions
    show_setup_instructions()

    print("=" * 60)
    print("✅ Setup completed! You're ready to use the examples.")
    print("=" * 60 + "\n")


if __name__ == "__main__":
    main()
