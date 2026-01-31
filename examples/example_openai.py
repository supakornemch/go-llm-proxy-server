#!/usr/bin/env python3
"""
OpenAI SDK Example - ใช้ OpenAI SDK ผ่าน Proxy
ที่ต้องเตรียม: Virtual Key จาก Proxy, Proxy Server ต้องรัน
"""

from openai import OpenAI

def main():
    # ตั้งค่า Client ให้ชี้มาที่ Proxy แทนที่ OpenAI โดยตรง
    client = OpenAI(
        api_key="vk-frontend-app",         # ★ เปลี่ยนเป็น Virtual Key จาก Proxy
        base_url="http://localhost:8132"   # ★ URL ของ Proxy Server
    )

    print("🤖 OpenAI via Proxy - Chat Completion Example\n")

    # ส่ง request ไปยัง Proxy
    response = client.chat.completions.create(
        model="gpt-4-turbo",               # ★ Alias ที่ตั้งไว้ตอน Assign
        messages=[
            {"role": "system", "content": "You are a helpful Thai assistant."},
            {"role": "user", "content": "สวัสดีค่ะ วันนี้อากาศเป็นอย่างไร"}
        ],
        temperature=0.7,
        max_tokens=256
    )

    # แสดงผลลัพธ์
    print(f"✅ Model: {response.model}")
    print(f"💬 Response: {response.choices[0].message.content}")
    print(f"📊 Usage: {response.usage.prompt_tokens} prompt tokens, {response.usage.completion_tokens} completion tokens")


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print(f"❌ Error: {e}")
        print("\n💡 Tips:")
        print("  1. ตรวจสอบว่า Proxy Server รันอยู่: ./llm-proxy serve")
        print("  2. ตรวจสอบ Virtual Key ถูกต้อง")
        print("  3. ติดตั้ง OpenAI SDK: pip install openai")
