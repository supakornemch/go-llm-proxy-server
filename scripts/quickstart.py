#!/usr/bin/env python3
"""
Integrated Quick Start - Auto setup via CLI + Execute Examples
Works like smoke_test.sh but with automatic Python execution
"""

import subprocess
import sys
import os
import re
import json
from pathlib import Path
from datetime import datetime

# ============ Configuration ============
PORT = 8132
DB_TYPE = os.getenv("DB_TYPE", "mongodb")
DB_DSN = os.getenv("DB_DSN", "mongodb://root:examplepassword@localhost:27017/llm_proxy?authSource=admin")
VKEY = f"vk-quickstart-{int(datetime.now().timestamp())}"

# Colors
GREEN = '\033[0;32m'
RED = '\033[0;31m'
YELLOW = '\033[1;33m'
BLUE = '\033[0;34m'
NC = '\033[0m'


class ProxySetup:
    """Auto-setup LLM Proxy: Connections, Models, Virtual Keys"""
    
    def __init__(self):
        self.vkey = VKEY
        self.vkey_id = None
        self.models = {}
    
    def log_step(self, msg):
        print(f"{BLUE}▶{NC} {msg}")
    
    def log_success(self, msg):
        print(f"{GREEN}✅{NC} {msg}")
    
    def log_error(self, msg):
        print(f"{RED}❌{NC} {msg}")
    
    def log_warning(self, msg):
        print(f"{YELLOW}⚠️{NC} {msg}")
    
    def run_cli_command(self, cmd):
        """Execute CLI command and capture output"""
        try:
            result = subprocess.run(
                cmd,
                shell=True,
                capture_output=True,
                text=True,
                timeout=10
            )
            return result.stdout + result.stderr
        except Exception as e:
            self.log_error(f"Command failed: {cmd}")
            self.log_error(str(e))
            return ""
    
    def extract_id(self, output):
        """Extract UUID from CLI output"""
        match = re.search(
            r'[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}',
            output,
            re.IGNORECASE
        )
        return match.group(0) if match else None
    
    def check_proxy_running(self):
        """Check if Proxy Server is running"""
        self.log_step("Checking Proxy Server...")
        try:
            result = subprocess.run(
                f"curl -s http://localhost:{PORT}/health",
                shell=True,
                capture_output=True,
                timeout=3
            )
            if result.returncode == 0:
                self.log_success("Proxy Server is running")
                return True
        except:
            pass
        
        self.log_error(f"Proxy Server is not running on port {PORT}")
        self.log_step("Start it with: ./llm-proxy serve")
        return False
    
    def create_virtual_key(self):
        """Create Virtual Key"""
        self.log_step("Creating Virtual Key...")
        
        cmd = f"./llm-proxy vkey add --db-type {DB_TYPE} --dsn \"{DB_DSN}\" --name \"QuickStart-{int(datetime.now().timestamp())}\" --key \"{self.vkey}\""
        output = self.run_cli_command(cmd)
        
        vkey_id = self.extract_id(output)
        if not vkey_id:
            self.log_error("Failed to create Virtual Key")
            return False
        
        self.vkey_id = vkey_id
        self.log_success(f"Virtual Key created: {vkey_id}")
        print(f"  Key: {self.vkey}")
        return True
    
    def create_connection(self, provider, name, endpoint, api_key):
        """Create Provider Connection"""
        if not api_key:
            self.log_warning(f"Skipping {name} (API Key not set)")
            return None
        
        self.log_step(f"Creating connection: {name}")
        
        cmd = f"""./llm-proxy connection add \\
            --db-type {DB_TYPE} \\
            --dsn "{DB_DSN}" \\
            --provider "{provider}" \\
            --name "{name}" \\
            --endpoint "{endpoint}" \\
            --api-key "{api_key}\""""
        
        output = self.run_cli_command(cmd)
        conn_id = self.extract_id(output)
        
        if not conn_id:
            self.log_error(f"Failed to create connection {name}")
            return None
        
        self.log_success(f"Connection created: {conn_id}")
        return conn_id
    
    def create_model(self, conn_id, model_name, remote_name):
        """Create Model in Connection"""
        if not conn_id:
            return None
        
        self.log_step(f"Adding model: {model_name} → {remote_name}")
        
        cmd = f"""./llm-proxy model add \\
            --db-type {DB_TYPE} \\
            --dsn "{DB_DSN}" \\
            --conn-id "{conn_id}" \\
            --name "{model_name}" \\
            --remote "{remote_name}\""""
        
        output = self.run_cli_command(cmd)
        model_id = self.extract_id(output)
        
        if not model_id:
            self.log_error(f"Failed to create model {model_name}")
            return None
        
        self.log_success(f"Model created: {model_id}")
        return model_id
    
    def assign_model(self, model_id, alias, tps=50):
        """Assign Model to Virtual Key"""
        if not model_id or not self.vkey_id:
            return False
        
        self.log_step(f"Assigning model: {alias} (TPS: {tps})")
        
        cmd = f"""./llm-proxy assign \\
            --db-type {DB_TYPE} \\
            --dsn "{DB_DSN}" \\
            --vkey-id "{self.vkey_id}" \\
            --model-id "{model_id}" \\
            --alias "{alias}" \\
            --tps {tps}"""
        
        output = self.run_cli_command(cmd)
        self.log_success(f"Model assigned: {alias}")
        self.models[alias] = model_id
        return True
    
    def setup_openai(self):
        """Setup OpenAI Provider"""
        self.log_step("🔴 Setting up OpenAI")
        
        api_key = os.getenv("OPENAI_API_KEY")
        endpoint = os.getenv("OPENAI_API_ENDPOINT", "https://api.openai.com")
        
        conn_id = self.create_connection("openai", "OpenAI-Main", endpoint, api_key)
        if conn_id:
            model_id = self.create_model(conn_id, "gpt-4-turbo", "gpt-4-turbo-preview")
            self.assign_model(model_id, "gpt-4-turbo")
    
    def setup_azure(self):
        """Setup Azure OpenAI Provider"""
        self.log_step("🔵 Setting up Azure OpenAI")
        
        api_key = os.getenv("AZURE_OPENAI_API_KEY")
        endpoint = os.getenv("AZURE_OPENAI_ENDPOINT")
        
        conn_id = self.create_connection("azure", "Azure-Main", endpoint, api_key)
        if conn_id:
            model_id = self.create_model(conn_id, "gpt-4o", "gpt-4o")
            self.assign_model(model_id, "gpt-4o")
    
    def setup_google(self):
        """Setup Google Vertex AI Provider"""
        self.log_step("🟢 Setting up Google Vertex AI")
        
        api_key = os.getenv("GOOGLE_VERTEX_API_KEY")
        endpoint = os.getenv("GOOGLE_GEMINI_ENDPOINT", "https://aiplatform.googleapis.com")
        
        conn_id = self.create_connection("google", "Google-Vertex", endpoint, api_key)
        if conn_id:
            model_id = self.create_model(conn_id, "gemini-3-flash", "gemini-3-flash-preview")
            self.assign_model(model_id, "gemini-3-flash-preview")
    
    def run_example(self, example_file, alias):
        """Execute Python Example"""
        example_path = Path("examples") / example_file
        
        if not example_path.exists():
            self.log_warning(f"Example file not found: {example_file}")
            return
        
        self.log_step(f"Running example: {example_file}")
        print(f"  Model: {alias}, Key: {self.vkey}")
        print()
        
        # Read example file
        with open(example_path, 'r') as f:
            code = f.read()
        
        # Replace placeholders
        code = code.replace('vk-frontend-app', self.vkey)
        code = code.replace('vk-google-app', self.vkey)
        code = code.replace('vk-azure-app', self.vkey)
        code = code.replace('vk-my-app', self.vkey)
        code = code.replace('gpt-4-turbo', alias)
        code = code.replace('gpt-4o', alias)
        code = code.replace('gemini-3-flash', alias)
        code = code.replace('gemini-2-flash', alias)
        
        # Execute
        try:
            import subprocess
            result = subprocess.run(
                ['./.venv/bin/python3', '-c', code],
                cwd='.',
                capture_output=False
            )
        except Exception as e:
            self.log_error(f"Example execution failed: {e}")
        
        print()
    
    def run(self, providers='all'):
        """Main execution"""
        print()
        print(f"{YELLOW}================================{NC}")
        print(f"{YELLOW}🚀 LLM Proxy Quick Start{NC}")
        print(f"{YELLOW}================================{NC}")
        print()
        
        # 1. Check Proxy
        if not self.check_proxy_running():
            return False
        print()
        
        # 2. Create Virtual Key
        if not self.create_virtual_key():
            return False
        print()
        
        # 3. Setup Providers
        if providers == 'openai' or providers == 'all':
            self.setup_openai()
            if providers == 'all':
                print()
        
        if providers == 'azure' or providers == 'all':
            self.setup_azure()
            if providers == 'all':
                print()
        
        if providers == 'google' or providers == 'all':
            self.setup_google()
            if providers == 'all':
                print()
        
        print(f"{GREEN}================================{NC}")
        print(f"{GREEN}✨ Setup Complete!{NC}")
        print(f"{GREEN}================================{NC}")
        print()
        print(f"Virtual Key: {self.vkey}")
        print(f"Base URL: http://localhost:{PORT}")
        print()
        
        # 4. Ask to run examples
        if self.models:
            try:
                response = input("Run Python examples now? (y/n): ").strip().lower()
                if response == 'y':
                    print()
                    
                    if 'gpt-4-turbo' in self.models:
                        self.run_example("example_openai.py", "gpt-4-turbo")
                    
                    if 'gpt-4o' in self.models:
                        self.run_example("example_azure.py", "gpt-4o")
                    
                    if 'gemini-3-flash-preview' in self.models:
                        self.run_example("example_google_vertex_http.py", "gemini-3-flash-preview")
                    
                    print(f"{GREEN}🎉 All examples executed!{NC}")
            except KeyboardInterrupt:
                print("\nSkipped")
        
        return True


def main():
    import argparse
    
    parser = argparse.ArgumentParser(
        description='LLM Proxy Quick Start - Auto setup and Python execution'
    )
    parser.add_argument(
        '--provider',
        choices=['openai', 'azure', 'google', 'all'],
        default='all',
        help='Provider to setup (default: all)'
    )
    parser.add_argument(
        '--skip-examples',
        action='store_true',
        help='Skip running Python examples'
    )
    
    args = parser.parse_args()
    
    setup = ProxySetup()
    success = setup.run(args.provider)
    
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
