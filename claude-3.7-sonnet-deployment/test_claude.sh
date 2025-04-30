#!/bin/bash

# Colors for terminal output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Testing Claude 3.7 Sonnet Deployment ===${NC}"

# Check if curl is installed
if ! command -v curl &> /dev/null; then
    echo -e "${RED}curl is not installed. Please install curl first.${NC}"
    exit 1
fi

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo -e "${YELLOW}jq is not installed. Output will not be formatted.${NC}"
    JQ_INSTALLED=false
else
    JQ_INSTALLED=true
fi

# Check if the service is running
if ! curl -s http://localhost:8080/v1/models > /dev/null; then
    echo -e "${RED}The chatgpt-adapter service is not running. Please start it first.${NC}"
    echo "Run: docker-compose up -d"
    exit 1
fi

echo -e "${GREEN}Service is running. Testing Claude 3.7 Sonnet...${NC}"

# Get token from file
if [ -f .cursor_token ]; then
    TOKEN=$(cat .cursor_token)
else
    echo -e "${YELLOW}No token file found. Using the API without authentication.${NC}"
    TOKEN=""
fi

# Set headers
if [ -n "$TOKEN" ]; then
    AUTH_HEADER="Authorization: Bearer $TOKEN"
else
    AUTH_HEADER=""
fi

# Make the API call
RESPONSE=$(curl -s -X POST http://localhost:8080/v1/chat/completions \
    -H "Content-Type: application/json" \
    -H "$AUTH_HEADER" \
    -d '{
        "model": "cursor/claude-3.7-sonnet",
        "messages": [{"role": "user", "content": "Hello, are you Claude 3.7 Sonnet? Please confirm your model name and provide a brief greeting."}]
    }')

# Check if the response is valid
if [ -z "$RESPONSE" ]; then
    echo -e "${RED}No response received from the API.${NC}"
    exit 1
fi

# Format and display the response
echo -e "${GREEN}Response received:${NC}"
if [ "$JQ_INSTALLED" = true ]; then
    echo "$RESPONSE" | jq .
else
    echo "$RESPONSE"
fi

# Extract the content from the response
if [ "$JQ_INSTALLED" = true ]; then
    CONTENT=$(echo "$RESPONSE" | jq -r '.choices[0].message.content')
    if [[ "$CONTENT" == *"Claude 3.7 Sonnet"* ]]; then
        echo -e "${GREEN}Success! Claude 3.7 Sonnet is working correctly.${NC}"
    else
        echo -e "${YELLOW}The response doesn't explicitly confirm Claude 3.7 Sonnet. Please check the response content.${NC}"
    fi
else
    echo -e "${YELLOW}Cannot verify model confirmation without jq. Please check the response manually.${NC}"
fi

