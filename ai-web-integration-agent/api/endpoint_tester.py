#!/usr/bin/env python3
"""
Streamlined script for testing Claude, Copilot, and Cursor endpoints.
This script allows you to:
1. Select which AI service to test (Claude, Copilot, or Cursor)
2. Specify the port for the API server
3. Test the endpoint with a simple "Hello" message
4. Shut down and restart with different options
"""

import argparse
import json
import os
import platform
import requests
import signal
import subprocess
import sys
import time
from typing import Dict, List, Optional, Tuple, Union

# ANSI color codes for terminal output
class Colors:
    HEADER = '\033[95m'
    BLUE = '\033[94m'
    GREEN = '\033[92m'
    YELLOW = '\033[93m'
    RED = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'

# Clear terminal screen based on OS
def clear_screen():
    if platform.system() == "Windows":
        os.system('cls')
    else:
        os.system('clear')

# Print colored text
def print_color(text: str, color: str):
    print(f"{color}{text}{Colors.ENDC}")

# Print header
def print_header(text: str):
    print("\n" + "=" * 80)
    print_color(f"  {text}", Colors.HEADER + Colors.BOLD)
    print("=" * 80 + "\n")

# Print status message
def print_status(text: str, success: bool = True):
    if success:
        print_color(f"✓ {text}", Colors.GREEN)
    else:
        print_color(f"✗ {text}", Colors.RED)

# Print info message
def print_info(text: str):
    print_color(f"ℹ {text}", Colors.BLUE)

# Print warning message
def print_warning(text: str):
    print_color(f"⚠ {text}", Colors.YELLOW)

# Service configurations
SERVICES = {
    "claude": {
        "name": "Claude",
        "model": "web_claude",
        "endpoint": "/v1/chat/completions",
        "request_template": {
            "model": "web_claude",
            "messages": [
                {"role": "user", "content": "Hello"}
            ],
            "temperature": 0.7,
            "stream": False
        }
    },
    "copilot": {
        "name": "GitHub Copilot",
        "model": "web_copilot",
        "endpoint": "/v1/completions",
        "request_template": {
            "model": "web_copilot",
            "prompt": "// Say hello\nfunction greet() {",
            "temperature": 0.7,
            "stream": False
        }
    },
    "cursor": {
        "name": "Cursor",
        "model": "cursor/claude-3.7-sonnet",
        "endpoint": "/v1/chat/completions",
        "request_template": {
            "model": "cursor/claude-3.7-sonnet",
            "messages": [
                {"role": "user", "content": "Hello"}
            ],
            "temperature": 0.7,
            "stream": False
        }
    }
}

# Server process
server_process = None

# Test if a port is in use
def is_port_in_use(port: int) -> bool:
    import socket
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        return s.connect_ex(('localhost', port)) == 0

# Start the API server
def start_server(port: int, service: str) -> subprocess.Popen:
    global server_process
    
    # Check if port is already in use
    if is_port_in_use(port):
        print_warning(f"Port {port} is already in use. This might be another instance of the API server.")
        response = input("Do you want to continue anyway? (y/n): ")
        if response.lower() != 'y':
            print_info("Exiting...")
            sys.exit(0)
    
    print_info(f"Starting API server for {SERVICES[service]['name']} on port {port}...")
    
    # Determine the correct path to the API server executable
    script_dir = os.path.dirname(os.path.abspath(__file__))
    server_path = os.path.join(script_dir, "web-integration-api")
    
    # Check if the server executable exists
    if not os.path.exists(server_path):
        # Try to build it
        print_info("Server executable not found. Attempting to build it...")
        try:
            build_process = subprocess.run(
                ["go", "build", "-o", "web-integration-api", "."],
                cwd=script_dir,
                check=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE
            )
            print_status("Server built successfully")
        except subprocess.CalledProcessError as e:
            print_status(f"Failed to build server: {e.stderr.decode()}", False)
            print_info("Please build the server manually using 'go build -o web-integration-api .'")
            sys.exit(1)
        except FileNotFoundError:
            print_status("Go compiler not found. Please install Go and build the server manually.", False)
            sys.exit(1)
    
    # Create a config file for the server
    config = {
        "port": port,
        "host": "0.0.0.0",
        "claude_url": "https://claude.ai/chat",
        "github_copilot_url": "https://github.com/features/copilot",
        "browser_user_data_dir": os.path.expanduser("~/.browser-agent"),
        "screenshot_dir": "./screenshots",
        "log_file": "./api-server.log",
        "headless": False,
        "debug_mode": True
    }
    
    config_path = os.path.join(script_dir, "config.json")
    with open(config_path, 'w') as f:
        json.dump(config, f, indent=2)
    
    # Start the server
    try:
        server_process = subprocess.Popen(
            [server_path, "--config", config_path, "--port", str(port)],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        
        # Give the server some time to start
        time.sleep(3)
        
        # Check if the server is running
        if server_process.poll() is not None:
            # Server has exited
            stdout, stderr = server_process.communicate()
            print_status(f"Server failed to start: {stderr}", False)
            return None
        
        print_status(f"API server started on port {port}")
        return server_process
    
    except Exception as e:
        print_status(f"Failed to start server: {str(e)}", False)
        return None

# Stop the API server
def stop_server():
    global server_process
    if server_process:
        print_info("Stopping API server...")
        
        # Try to terminate gracefully first
        if platform.system() == "Windows":
            server_process.terminate()
        else:
            server_process.send_signal(signal.SIGTERM)
        
        # Wait for a bit to see if it terminates
        try:
            server_process.wait(timeout=5)
            print_status("API server stopped")
        except subprocess.TimeoutExpired:
            # Force kill if it doesn't terminate
            print_warning("Server not responding to termination signal. Force killing...")
            if platform.system() == "Windows":
                server_process.kill()
            else:
                server_process.send_signal(signal.SIGKILL)
            print_status("API server killed")
        
        server_process = None

# Test the API endpoint
def test_endpoint(port: int, service: str) -> bool:
    print_info(f"Testing {SERVICES[service]['name']} endpoint...")
    
    url = f"http://localhost:{port}{SERVICES[service]['endpoint']}"
    headers = {
        "Content-Type": "application/json",
        "Accept": "application/json"
    }
    
    try:
        response = requests.post(
            url,
            headers=headers,
            json=SERVICES[service]['request_template'],
            timeout=30
        )
        
        if response.status_code == 200:
            print_status(f"{SERVICES[service]['name']} endpoint is working!")
            print_info("Response:")
            
            # Pretty print the response
            try:
                response_json = response.json()
                print(json.dumps(response_json, indent=2))
                return True
            except json.JSONDecodeError:
                print(response.text)
                return True
        else:
            print_status(f"Endpoint test failed with status code: {response.status_code}", False)
            print_info("Response:")
            print(response.text)
            return False
    
    except requests.exceptions.ConnectionError:
        print_status("Connection error. Make sure the server is running.", False)
        return False
    except requests.exceptions.Timeout:
        print_status("Request timed out. The server might be busy or not responding.", False)
        return False
    except Exception as e:
        print_status(f"Error testing endpoint: {str(e)}", False)
        return False

# Main menu
def main_menu() -> Tuple[str, int]:
    clear_screen()
    print_header("AI Web Integration API Tester")
    
    print("Select an AI service to test:")
    print("1. Claude")
    print("2. GitHub Copilot")
    print("3. Cursor")
    print("4. Exit")
    
    while True:
        choice = input("\nEnter your choice (1-4): ")
        if choice == "1":
            service = "claude"
            break
        elif choice == "2":
            service = "copilot"
            break
        elif choice == "3":
            service = "cursor"
            break
        elif choice == "4":
            print_info("Exiting...")
            sys.exit(0)
        else:
            print_warning("Invalid choice. Please try again.")
    
    # Get port number
    while True:
        port_input = input("\nEnter port number (default: 8080): ")
        if port_input == "":
            port = 8080
            break
        
        try:
            port = int(port_input)
            if 1024 <= port <= 65535:
                break
            else:
                print_warning("Port must be between 1024 and 65535.")
        except ValueError:
            print_warning("Invalid port number. Please enter a valid integer.")
    
    return service, port

# Main function
def main():
    try:
        while True:
            # Show menu and get choices
            service, port = main_menu()
            
            # Start server
            server = start_server(port, service)
            if not server:
                input("\nPress Enter to continue...")
                continue
            
            # Test endpoint
            test_result = test_endpoint(port, service)
            
            # Ask what to do next
            print("\nWhat would you like to do next?")
            print("1. Test another service")
            print("2. Exit")
            
            while True:
                choice = input("\nEnter your choice (1-2): ")
                if choice == "1" or choice == "2":
                    break
                else:
                    print_warning("Invalid choice. Please try again.")
            
            # Stop the current server
            stop_server()
            
            if choice == "2":
                print_info("Exiting...")
                break
    
    except KeyboardInterrupt:
        print_info("\nInterrupted by user. Cleaning up...")
    finally:
        # Make sure to stop the server when exiting
        stop_server()

if __name__ == "__main__":
    main()

