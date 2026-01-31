#!/usr/bin/env python3
"""
AWS Bedrock Example - ใช้ boto3 ผ่าน Proxy
ที่ต้องเตรียม: Virtual Key จาก Proxy, Proxy Server ต้องรัน
Note: นี่เป็นตัวอย่าง - AWS Bedrock อาจต้องการการตั้งค่าพิเศษ
"""

import json
import boto3
from botocore.config import Config

def main():
    print("🏗️  AWS Bedrock via Proxy - Example\n")

    # ★ สำคัญ: AWS Bedrock ต้องการการตั้งค่า Credentials
    # หากใช้ผ่าน Proxy อาจต้องเปลี่ยน endpoint_url เป็น Proxy URL
    
    # ตัวอย่างการตั้งค่า (ทั่วไป AWS)
    bedrock_client = boto3.client(
        'bedrock-runtime',
        region_name='us-east-1',
        config=Config(
            retries={'max_attempts': 2},
            connect_timeout=10,
            read_timeout=60
        )
    )

    print("💡 Note: AWS Bedrock integration with Proxy requires special setup")
    print("   Please configure AWS credentials and endpoint URL accordingly\n")

    # ตัวอย่าง payload สำหรับ Claude model
    payload = {
        "prompt": "\n\nHuman: Explain quantum computing in simple terms\n\nAssistant:",
        "temperature": 0.7,
        "max_tokens_to_sample": 512
    }

    try:
        # สำหรับการทดสอบจริง ต้องมี Bedrock access
        print("📤 Sending request to Bedrock (via Proxy)...")
        print(f"   Payload: {json.dumps(payload, indent=2)}\n")

        # Uncomment เมื่อพร้อมใช้จริง
        # response = bedrock_client.invoke_model(
        #     modelId='anthropic.claude-3-sonnet-20240229-v1:0',
        #     body=json.dumps(payload)
        # )
        #
        # output = json.loads(response['body'].read())
        # print(f"✅ Response: {output['completion']}")

        print("✅ Setup completed. Ready to invoke Bedrock models.")
        print("   Uncomment the invoke_model call to run with real Bedrock access.\n")

    except Exception as e:
        print(f"❌ Error: {e}")
        print("\n💡 Tips:")
        print("  1. ตรวจสอบ AWS credentials ถูกต้อง")
        print("  2. ติดตั้ง boto3: pip install boto3")
        print("  3. Proxy ต้องมีการตั้งค่า Bedrock endpoint เพิ่มเติม")


if __name__ == "__main__":
    main()
