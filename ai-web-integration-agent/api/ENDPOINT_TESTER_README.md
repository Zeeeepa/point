# AI Web Integration API Endpoint Tester

This directory contains scripts for testing the AI Web Integration API endpoints for Claude, GitHub Copilot, and Cursor.

## Features

- Select which AI service to test (Claude, Copilot, or Cursor)
- Specify the port for the API server
- Test the endpoint with a simple "Hello" message
- Shut down and restart with different options

## Available Scripts

Choose the script that best fits your operating system:

- `endpoint_tester.py` - Python script (cross-platform)
- `endpoint_tester.sh` - Bash script (Linux/macOS)
- `endpoint_tester.bat` - Batch script (Windows)

## Prerequisites

### For Python Script

- Python 3.6 or higher
- `requests` library (install with `pip install requests`)

### For Bash Script

- Bash shell
- `curl` command-line tool
- `jq` command-line tool (optional, for pretty-printing JSON)

### For Batch Script

- Windows command prompt
- `curl` command-line tool

## Usage

### Python Script

```bash
# Make the script executable (Linux/macOS)
chmod +x endpoint_tester.py

# Run the script
./endpoint_tester.py
```

Or:

```bash
python endpoint_tester.py
```

### Bash Script

```bash
# Make the script executable
chmod +x endpoint_tester.sh

# Run the script
./endpoint_tester.sh
```

### Batch Script

```cmd
# Run the script
endpoint_tester.bat
```

## How It Works

1. The script displays a menu to select which AI service to test (Claude, Copilot, or Cursor)
2. You can specify the port for the API server (default: 8080)
3. The script starts the API server with the appropriate configuration
4. It sends a test message ("Hello") to the selected endpoint
5. The response is displayed
6. You can choose to test another service or exit

## Troubleshooting

### Server Fails to Start

- Make sure the Go executable is in your PATH
- Make sure you have the necessary dependencies installed
- Check the API server logs in `api-server.log`

### Connection Errors

- Make sure the port is not already in use
- Check if your firewall is blocking the connection
- Verify that the API server is running

### Authentication Errors

- Make sure you're logged in to the respective service (Claude, GitHub, etc.)
- Check the browser automation logs

## Configuration

The script creates a `config.json` file with the following default settings:

```json
{
  "port": 8080,
  "host": "0.0.0.0",
  "claude_url": "https://claude.ai/chat",
  "github_copilot_url": "https://github.com/features/copilot",
  "browser_user_data_dir": "~/.browser-agent",
  "screenshot_dir": "./screenshots",
  "log_file": "./api-server.log",
  "headless": false,
  "debug_mode": true
}
```

You can modify this file to change the default settings.

