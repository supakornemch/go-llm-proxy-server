#!/usr/bin/env python3
"""
Google Generative AI SDK Example - ใช้ Google Gemini SDK ผ่าน Proxy
ที่ต้องเตรียม: Virtual Key จาก Proxy, Proxy Server ต้องรัน
"""

import google.generativeai as genai

def main():
    # ตั้งค่า Vertex AI SDK ให้ใช้ Virtual Key
    genai.configure(api_key="vk-vertex-app")  # ★ Virtual Key จาก Proxy

    print("🔍 Google Gemini via Proxy - Text Generation Example\n")

    # สร้าง Model instance
    model = genai.GenerativeModel(
        model_name="gemini-3-flash"         # ★ Alias ที่ตั้งไว้ตอน Assign
    )

    # ส่ง request
    response = model.generate_content(
        "อธิบายความเป็นมา AI และ Machine Learning แบบง่าย ๆ"
    )

    # แสดงผลลัพธ์
    print(f"💬 Response:\n{response.text}")
    print(f"\n✅ Generation completed successfully")


if __name__ == "__main__":
    try:
        main()
    except Exception as e:
        print(f"❌ Error: {e}")
        print("\n💡 Tips:")
        print("  1. ตรวจสอบว่า Proxy Server รันอยู่: ./llm-proxy serve")
        print("  2. ตรวจสอบ Virtual Key ถูกต้อง")
        print("  3. ติดตั้ง Google Generative AI SDK: pip install google-generativeai")
