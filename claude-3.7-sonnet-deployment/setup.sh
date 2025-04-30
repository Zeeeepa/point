#!/bin/bash

# Colors for terminal output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Claude 3.7 Sonnet Deployment Setup ===${NC}"
echo -e "${YELLOW}This script will help you set up the chatgpt-adapter for Claude 3.7 Sonnet via Cursor${NC}"
echo ""

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker is not installed. Please install Docker first.${NC}"
    echo "Visit https://docs.docker.com/get-docker/ for installation instructions."
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}Docker Compose is not installed. Please install Docker Compose first.${NC}"
    echo "Visit https://docs.docker.com/compose/install/ for installation instructions."
    exit 1
fi

# Get Cursor session token
echo -e "${YELLOW}You need a valid Cursor session token (WorkosCursorSessionToken) to use Claude 3.7 Sonnet.${NC}"
echo -e "${YELLOW}You can get this token by logging into Cursor (https://www.cursor.com) and extracting it from your browser cookies.${NC}"
echo ""
echo -e "${BLUE}To get your token:${NC}"
echo "1. Log in to Cursor (https://www.cursor.com)"
echo "2. Open your browser's developer tools (F12 or right-click > Inspect)"
echo "3. Go to the Application/Storage tab"
echo "4. Find Cookies > https://www.cursor.com"
echo "5. Copy the value of the 'WorkosCursorSessionToken' cookie"
echo ""

read -p "Enter your Cursor session token (WorkosCursorSessionToken): " CURSOR_TOKEN

if [ -z "$CURSOR_TOKEN" ]; then
    echo -e "${RED}No token provided. Setup cannot continue.${NC}"
    exit 1
fi

# Create a token file
echo -e "${GREEN}Saving token...${NC}"
echo "$CURSOR_TOKEN" > .cursor_token

# Start the service
echo -e "${GREEN}Starting chatgpt-adapter with Claude 3.7 Sonnet support...${NC}"
docker-compose up -d

echo ""
echo -e "${GREEN}=== Setup Complete! ===${NC}"
echo -e "${BLUE}The chatgpt-adapter is now running at http://localhost:8080${NC}"
echo ""
echo -e "${YELLOW}Available Claude 3.7 Models:${NC}"
echo "- claude-3.7-sonnet"
echo "- claude-3.7-sonnet-max"
echo "- claude-3.7-sonnet-thinking"
echo "- claude-3.7-sonnet-thinking-max"
echo ""
echo -e "${BLUE}To test the deployment, run:${NC}"
echo "./test_claude.sh"
echo ""
echo -e "${YELLOW}To stop the service:${NC}"
echo "docker-compose down"

