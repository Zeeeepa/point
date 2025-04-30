# Claude 3.7 Sonnet Deployment

This directory contains everything you need to deploy Claude 3.7 Sonnet via the chatgpt-adapter, which allows you to use Claude 3.7 Sonnet through the standard OpenAI API format.

## Available Claude 3.7 Models

- `claude-3.7-sonnet` - Standard Claude 3.7 Sonnet model
- `claude-3.7-sonnet-max` - Claude 3.7 Sonnet with extended context window
- `claude-3.7-sonnet-thinking` - Claude 3.7 Sonnet with thinking capabilities
- `claude-3.7-sonnet-thinking-max` - Claude 3.7 Sonnet with thinking capabilities and extended context window

## Prerequisites

- Docker and Docker Compose
- A valid Cursor account with access to Claude 3.7 Sonnet
- curl (for testing)
- jq (optional, for better test output formatting)

## Setup Instructions

### Linux/macOS

1. Clone this repository:
   ```bash
   git clone https://github.com/Zeeeepa/point.git
   cd point/claude-3.7-sonnet-deployment
   ```

2. Make the scripts executable:
   ```bash
   chmod +x setup.sh test_claude.sh
   ```

3. Run the setup script:
   ```bash
   ./setup.sh
   ```

4. Follow the prompts to enter your Cursor session token.

5. Test the deployment:
   ```bash
   ./test_claude.sh
   ```

### Windows

1. Clone this repository:
   ```powershell
   git clone https://github.com/Zeeeepa/point.git
   cd point/claude-3.7-sonnet-deployment
   ```

2. Run one of the Windows setup scripts:

   **Option 1: Using Command Prompt (CMD):**
   ```cmd
   setup.bat
   ```

   **Option 2: Using PowerShell (recommended):**
   ```powershell
   .\setup.ps1
   ```

3. Follow the prompts to enter your Cursor session token.

4. Test the deployment:

   **Option 1: Using Command Prompt (CMD):**
   ```cmd
   test_claude.bat
   ```

   **Option 2: Using PowerShell (recommended):**
   ```powershell
   .\test_claude.ps1
   ```

## Getting Your Cursor Session Token

You need a valid Cursor session token (WorkosCursorSessionToken) to use Claude 3.7 Sonnet. Here's how to get it:

### Method 1: Using Developer Tools

1. Log in to Cursor (https://www.cursor.com)
2. Open your browser's developer tools (F12 or right-click > Inspect)
3. Go to the Application/Storage tab
4. Find Cookies > https://www.cursor.com
5. Copy the value of the 'WorkosCursorSessionToken' cookie

### Method 2: Using the Bookmarklet

1. Create a new bookmark in your browser
2. Name it "Get Cursor Token"
3. Paste the contents of `get_cursor_token.js` as the URL
4. While logged into Cursor, click the bookmark
5. A prompt will appear with your token

## Using Claude 3.7 Sonnet

After deployment, you can make requests to Claude 3.7 Sonnet using the standard OpenAI API format:

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "cursor/claude-3.7-sonnet",
    "messages": [{"role": "user", "content": "Hello, how are you?"}]
  }'
```

For Windows PowerShell:

```powershell
Invoke-RestMethod -Uri "http://localhost:8080/v1/chat/completions" `
  -Method Post `
  -Headers @{"Content-Type"="application/json"} `
  -Body '{
    "model": "cursor/claude-3.7-sonnet",
    "messages": [{"role": "user", "content": "Hello, how are you?"}]
  }'
```

For Windows Command Prompt:

```cmd
curl -X POST http://localhost:8080/v1/chat/completions ^
  -H "Content-Type: application/json" ^
  -d "{\"model\":\"cursor/claude-3.7-sonnet\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello, how are you?\"}]}"
```

## API Endpoints

The chatgpt-adapter provides the following OpenAI-compatible endpoints:

- `/v1/chat/completions` - For chat completions
- `/v1/models` - To list available models
- `/v1/embeddings` - For embeddings (if supported)
- `/v1/images/generations` - For image generation (if supported)

## Stopping the Service

To stop the service:

```bash
docker-compose down
```

## Troubleshooting

- **No response from API**: Check if your Cursor session token is valid and not expired
- **Error in response**: Check the error message for details
- **Service not starting**: Check Docker logs with `docker-compose logs`
- **Windows-specific issues**: 
  - If you encounter permission issues, try running the command prompt or PowerShell as Administrator
  - If Docker Desktop isn't running, start it before running the setup script
  - For WSL2 backend issues, ensure WSL2 is properly configured
  - If you get "The system cannot find the path specified" errors, make sure you're in the correct directory

## Security Notes

- Your Cursor session token is stored locally in the `.cursor_token` file
- Consider adding this file to your `.gitignore` to prevent accidental commits
- The token grants access to your Cursor account, so keep it secure

## Credits

This deployment uses the [chatgpt-adapter](https://github.com/bincooo/chatgpt-adapter) project, which provides a unified interface to various AI services.
