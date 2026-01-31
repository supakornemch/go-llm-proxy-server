#!/usr/bin/env python3
"""
Azure OpenAI SDK Example - ใช้ Azure OpenAI SDK ผ่าน Proxy
ที่ต้องเตรียม: Virtual Key จาก Proxy, Proxy Server ต้องรัน
"""

from openai import AzureOpenAI

def main():
    # ทำให้ Azure SDK ชี้มาที่ Proxy แทนที่ Azure โดยตรง
    client = AzureOpenAI(
        api_key="vk-azure-app",            # ★ Virtual Key จาก Proxy
        api_version="2024-05-01-preview",  # Proxy จะจัดการให้
        base_url="http://localhost:8132"   # ★ URL ของ Proxy Server
    )

    print("☁️  Azure OpenAI via Proxy - Chat Completion Example\n")

    # ส่ง request ไปยัง Proxy
    response = client.chat.completions.create(
        model="gpt-4o",                    # ★ Alias ที่ตั้งไว้ตอน Assign
        messages=[
            {"role": "system", "content": "You are a code expert."},
            {"role": "user", "content": "เขียน Python function ที่บวกตัวเลข 2 ตัว"}
        ],
        temperature=0.5,
        max_tokens=512
    )

    # แสดงผลลัพธ์
    print(f"✅ Model: {response.model}")
    print(f"💬 Response:\n{response.choices[0].message.content}")
    print(f"📊 Usage: {response.usage.prompt_tokens} prompt tokens, {response.usage.completion_tokens} completion tokens")


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print(f"❌ Error: {e}")
        print("\n💡 Tips:")
        print("  1. ตรวจสอบว่า Proxy Server รันอยู่: ./llm-proxy serve")
        print("  2. ตรวจสอบ Virtual Key ถูกต้อง")
        print("  3. ติดตั้ง Azure OpenAI SDK: pip install azure-openai")
