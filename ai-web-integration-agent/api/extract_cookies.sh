#!/bin/bash

# Script to extract cookies from browsers for Claude and GitHub

SERVICE=$1

if [ -z "$SERVICE" ]; then
    echo "Usage: $0 <service>"
    echo "  service: claude or github"
    exit 1
fi

COOKIE_FILE=""
DOMAIN=""

if [ "$SERVICE" == "claude" ]; then
    COOKIE_FILE="claude_cookies.json"
    DOMAIN="claude.ai"
    echo "Extracting cookies for Claude..."
elif [ "$SERVICE" == "github" ]; then
    COOKIE_FILE="github_cookies.json"
    DOMAIN="github.com"
    echo "Extracting cookies for GitHub..."
else
    echo "Unknown service: $SERVICE"
    echo "Supported services: claude, github"
    exit 1
fi

# Check if Chrome is installed
CHROME_PATH=""
if [ -f "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" ]; then
    # macOS
    CHROME_PATH="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
elif [ -f "/usr/bin/google-chrome" ]; then
    # Linux
    CHROME_PATH="/usr/bin/google-chrome"
elif [ -f "/c/Program Files/Google/Chrome/Application/chrome.exe" ]; then
    # Windows
    CHROME_PATH="/c/Program Files/Google/Chrome/Application/chrome.exe"
elif [ -f "/c/Program Files (x86)/Google/Chrome/Application/chrome.exe" ]; then
    # Windows (x86)
    CHROME_PATH="/c/Program Files (x86)/Google/Chrome/Application/chrome.exe"
else
    echo "Chrome not found. Please install Chrome or specify the path manually."
    exit 1
fi

echo "Using Chrome at: $CHROME_PATH"

# Create a temporary profile directory
TEMP_DIR=$(mktemp -d)
echo "Created temporary profile directory: $TEMP_DIR"

# Open Chrome with the service's website
echo "Opening Chrome to log in to $SERVICE..."
"$CHROME_PATH" --user-data-dir="$TEMP_DIR" --no-first-run "https://$DOMAIN" &
CHROME_PID=$!

# Wait for user to log in
echo "Please log in to $SERVICE in the opened browser window."
echo "Press Enter when you're done..."
read

# Kill Chrome
kill $CHROME_PID
sleep 2

# Extract cookies
echo "Extracting cookies..."
COOKIES_DB="$TEMP_DIR/Default/Cookies"

if [ -f "$COOKIES_DB" ]; then
    # Create a copy of the database to avoid locking issues
    cp "$COOKIES_DB" "$COOKIES_DB.copy"
    
    # Extract cookies using sqlite3
    if command -v sqlite3 >/dev/null 2>&1; then
        echo "[" > "$COOKIE_FILE"
        sqlite3 -json "$COOKIES_DB.copy" "SELECT name, value, host_key, path, expires_utc, is_secure, is_httponly FROM cookies WHERE host_key LIKE '%$DOMAIN%';" | sed 's/\[\]//' >> "$COOKIE_FILE"
        echo "]" >> "$COOKIE_FILE"
        echo "Cookies extracted to $COOKIE_FILE"
    else
        echo "sqlite3 not found. Please install sqlite3 to extract cookies."
        exit 1
    fi
else
    echo "Cookies database not found at $COOKIES_DB"
    exit 1
fi

# Clean up
rm -rf "$TEMP_DIR"
echo "Temporary profile directory removed."

echo "Done!"

