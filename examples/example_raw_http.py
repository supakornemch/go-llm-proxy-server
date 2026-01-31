#!/usr/bin/env python3
"""
Raw HTTP Request Example - Universal approach ที่ใช้ได้กับทุก Provider
ที่ต้องเตรียม: Virtual Key จาก Proxy, Proxy Server ต้องรัน
"""

import requests
import json

def send_openai_request():
    """ส่ง OpenAI-compatible request ไปยัง Proxy"""
    url = "http://localhost:8132/v1/chat/completions"

    headers = {
        "Authorization": "Bearer vk-my-app",  # ★ Virtual Key จาก Proxy
        "Content-Type": "application/json"
    }

    payload = {
        "model": "gpt-4-turbo",  # ★ Alias ที่ตั้งไว้ตอน Assign
        "messages": [
            {"role": "system", "content": "You are a helpful assistant."},
            {"role": "user", "content": "Hello, can you help me?"}
        ],
        "temperature": 0.7,
        "max_tokens": 256
    }

    print("📤 OpenAI-Compatible Request Example\n")

    try:
        response = requests.post(url, headers=headers, json=payload)

        if response.status_code == 200:
            data = response.json()
            print(f"✅ Status: {response.status_code}")
            print(f"💬 Response: {data['choices'][0]['message']['content']}")
            print(f"📊 Usage: {data['usage']['prompt_tokens']} → {data['usage']['completion_tokens']} tokens\n")
        else:
            print(f"❌ Error {response.status_code}: {response.text}\n")

    except requests.exceptions.ConnectionError:
        print("❌ Connection Error: ไม่สามารถเชื่อมต่อ Proxy Server\n")
    except Exception as e:
        print(f"❌ Error: {e}\n")


def send_google_request():
    """ส่ง Google Vertex/Gemini request ไปยัง Proxy"""
    url = "http://localhost:8132/v1/publishers/google/models/gemini-2-flash:generateContent"

    headers = {
        "Authorization": "Bearer vk-google-app",  # ★ Virtual Key จาก Proxy
        "Content-Type": "application/json"
    }

    payload = {
        "contents": [
            {
                "role": "user",
                "parts": [
                    {"text": "What is the capital of Thailand?"}
                ]
            }
        ]
    }

    print("📤 Google Vertex AI Request Example\n")

    try:
        response = requests.post(url, headers=headers, json=payload)

        if response.status_code == 200:
            data = response.json()
            print(f"✅ Status: {response.status_code}")
            if "candidates" in data and len(data["candidates"]) > 0:
                text = data["candidates"][0]["content"]["parts"][0]["text"]
                print(f"💬 Response: {text}\n")
            else:
                print(f"Response: {json.dumps(data, indent=2)}\n")
        else:
            print(f"❌ Error {response.status_code}: {response.text}\n")

    except requests.exceptions.ConnectionError:
        print("❌ Connection Error: ไม่สามารถเชื่อมต่อ Proxy Server\n")
    except Exception as e:
        print(f"❌ Error: {e}\n")


def send_azure_request():
    """ส่ง Azure OpenAI request ไปยัง Proxy"""
    url = "http://localhost:8132/v1/chat/completions"

    headers = {
        "Authorization": "Bearer vk-azure-app",  # ★ Virtual Key จาก Proxy
        "Content-Type": "application/json"
    }

    payload = {
        "model": "gpt-4o",  # ★ Azure Deployment Name
        "messages": [
            {"role": "user", "content": "สวัสดี"}
        ]
    }

    print("📤 Azure OpenAI Request Example\n")

    try:
        response = requests.post(url, headers=headers, json=payload)

        if response.status_code == 200:
            data = response.json()
            print(f"✅ Status: {response.status_code}")
            print(f"💬 Response: {data['choices'][0]['message']['content']}\n")
        else:
            print(f"❌ Error {response.status_code}: {response.text}\n")

    except requests.exceptions.ConnectionError:
        print("❌ Connection Error: ไม่สามารถเชื่อมต่อ Proxy Server\n")
    except Exception as e:
        print(f"❌ Error: {e}\n")


def main():
    print("=" * 60)
    print("🌐 Raw HTTP Request Examples - All Providers")
    print("=" * 60 + "\n")

    # เลือกตัวอย่างที่จะรัน
    print("Choose an example to run:")
    print("1. OpenAI-compatible request")
    print("2. Google Vertex AI request")
    print("3. Azure OpenAI request")
    print("4. Run all examples")

    choice = input("\nEnter your choice (1-4): ").strip()

    if choice == "1":
        send_openai_request()
    elif choice == "2":
        send_google_request()
    elif choice == "3":
        send_azure_request()
    elif choice == "4":
        send_openai_request()
        send_google_request()
        send_azure_request()
    else:
        print("Invalid choice")


if __name__ == "__main__":
    main()
