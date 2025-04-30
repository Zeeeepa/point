#!/usr/bin/env python3

import argparse
import json
import requests
import sseclient
import sys

def main():
    parser = argparse.ArgumentParser(description='Test the AI Web Integration API Server')
    parser.add_argument('--url', default='http://localhost:8080', help='API server URL')
    parser.add_argument('--model', default='web_claude', choices=['web_claude', 'web_claude/chat', 'web_copilot', 'web_copilot/github'], help='Model to use')
    parser.add_argument('--prompt', default='Hello, how are you?', help='Prompt to send')
    parser.add_argument('--stream', action='store_true', help='Use streaming API')
    args = parser.parse_args()

    # Determine endpoint based on model
    if args.model.startswith('web_claude'):
        endpoint = f"{args.url}/v1/chat/completions"
        data = {
            "model": args.model,
            "messages": [
                {"role": "user", "content": args.prompt}
            ],
            "stream": args.stream
        }
    else:
        endpoint = f"{args.url}/v1/completions"
        data = {
            "model": args.model,
            "prompt": args.prompt,
            "stream": args.stream
        }

    headers = {
        "Content-Type": "application/json",
        "Accept": "application/json"
    }

    print(f"Sending request to {endpoint}...")
    print(f"Request data: {json.dumps(data, indent=2)}")

    if args.stream:
        # Handle streaming response
        response = requests.post(endpoint, json=data, headers=headers, stream=True)
        
        if response.status_code != 200:
            print(f"Error: {response.status_code} {response.text}")
            sys.exit(1)
        
        client = sseclient.SSEClient(response)
        
        for event in client.events():
            if event.data == "[DONE]":
                break
            
            try:
                chunk = json.loads(event.data)
                
                if "choices" in chunk:
                    if args.model.startswith('web_claude'):
                        if "delta" in chunk["choices"][0] and "content" in chunk["choices"][0]["delta"]:
                            content = chunk["choices"][0]["delta"]["content"]
                            print(content, end="", flush=True)
                    else:
                        if "text" in chunk["choices"][0]:
                            text = chunk["choices"][0]["text"]
                            print(text, end="", flush=True)
            except json.JSONDecodeError:
                print(f"Error parsing JSON: {event.data}")
        
        print("\n")
    else:
        # Handle non-streaming response
        response = requests.post(endpoint, json=data, headers=headers)
        
        if response.status_code != 200:
            print(f"Error: {response.status_code} {response.text}")
            sys.exit(1)
        
        result = response.json()
        
        if args.model.startswith('web_claude'):
            content = result["choices"][0]["message"]["content"]
            print(f"Response: {content}")
        else:
            text = result["choices"][0]["text"]
            print(f"Response: {text}")

if __name__ == "__main__":
    main()

