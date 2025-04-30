#!/bin/bash

# Script to test the integration between the API server and chatgpt-adapter-main

# Check if the API server is running
echo "Checking if the API server is running..."
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "API server is not running. Starting it..."
    cd "$(dirname "$0")"
    ./web-integration-api &
    API_SERVER_PID=$!
    
    # Wait for the API server to start
    echo "Waiting for the API server to start..."
    for i in {1..10}; do
        if curl -s http://localhost:8080/health > /dev/null; then
            echo "API server started successfully."
            break
        fi
        
        if [ $i -eq 10 ]; then
            echo "Failed to start API server."
            exit 1
        fi
        
        sleep 1
    done
else
    echo "API server is already running."
fi

# Test Claude endpoint
echo "Testing Claude endpoint..."
curl -s -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "web_claude",
    "messages": [
      {"role": "user", "content": "Say hello in 5 different languages."}
    ],
    "stream": false
  }' | jq .

# Test Copilot endpoint
echo "Testing Copilot endpoint..."
curl -s -X POST http://localhost:8080/v1/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "web_copilot",
    "prompt": "function calculateSum(a, b) {",
    "stream": false
  }' | jq .

# Clean up
if [ -n "$API_SERVER_PID" ]; then
    echo "Stopping API server..."
    kill $API_SERVER_PID
fi

echo "Done!"

