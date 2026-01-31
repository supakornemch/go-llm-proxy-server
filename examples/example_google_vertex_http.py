#!/usr/bin/env python3
"""
Google Vertex AI via HTTP - ส่ง Native JSON Request ไปยัง Proxy
รองรับ Thinking Config สำหรับ Gemini 3.0
ที่ต้องเตรียม: Virtual Key จาก Proxy, Proxy Server ต้องรัน
"""

import requests
import json

def main():
    # URL ของ Proxy Server
    url = "http://localhost:8132/v1/publishers/google/models/gemini-3-flash:generateContent"

    # Headers
    headers = {
        "Authorization": "Bearer vk-vertex-app",  # ★ Virtual Key จาก Proxy
        "Content-Type": "application/json"
    }

    # Native Google JSON Payload
    payload = {
        "contents": [
            {
                "role": "user",
                "parts": [
                    {"text": "เขียน Python function ที่หาตัวเลขที่ใหญ่ที่สุด"}
                ]
            }
        ],
        "generationConfig": {
            "temperature": 0.8,
            "maxOutputTokens": 1024,
            "topP": 0.95
        },
        # ★ Thinking Config - สำหรับ Gemini 3.0
        "thinkingConfig": {
            "type": "EXTENDED_THINKING",
            "budgetTokens": 5000
        }
    }

    print("🔍 Google Vertex AI via HTTP - Native JSON Example\n")
    print("📤 Sending request to Proxy...\n")

    try:
        # ส่ง POST request
        response = requests.post(url, headers=headers, json=payload)

        if response.status_code == 200:
            result = response.json()

            # แสดงผลลัพธ์
            print(f"✅ Status Code: {response.status_code}")
            
            if "candidates" in result and len(result["candidates"]) > 0:
                candidate = result["candidates"][0]
                
                # แสดง Thinking content ถ้ามี
                if "content" in candidate:
                    for part in candidate["content"].get("parts", []):
                        if "thinkingNote" in part:
                            print(f"\n💭 Extended Thinking:\n{part['thinkingNote']}\n")
                        if "text" in part:
                            print(f"💬 Response:\n{part['text']}")

                # แสดง Usage
                if "usageMetadata" in candidate:
                    usage = candidate["usageMetadata"]
                    print(f"\n📊 Usage:")
                    print(f"   - Prompt Tokens: {usage.get('promptTokenCount', 0)}")
                    print(f"   - Candidate Tokens: {usage.get('candidatesTokenCount', 0)}")
            else:
                print("No candidates in response")
                print(f"Full response: {json.dumps(result, indent=2)}")
        else:
            print(f"❌ Error {response.status_code}")
            print(f"Response: {response.text}")

    except requests.exceptions.ConnectionError:
        print("❌ Connection Error: ไม่สามารถเชื่อมต่อ Proxy Server")
        print("💡 ตรวจสอบว่า Proxy Server รันอยู่: ./llm-proxy serve")
    except Exception as e:
        print(f"❌ Error: {e}")


if __name__ == "__main__":
    main()
