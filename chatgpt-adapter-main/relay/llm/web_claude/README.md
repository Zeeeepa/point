# Web Claude Adapter

This adapter provides OpenAI API compatibility for Claude AI through browser automation. It allows you to interact with Claude's web interface without using their official API.

## Features

- Browser automation to interact with Claude's web interface
- Conversation history support
- Streaming responses
- OpenAI API compatibility
- System, user, and assistant message support
- Interactive login process

## Requirements

- [Playwright](https://playwright.dev/) for browser automation
- Valid Claude.ai cookies for authentication (can be generated through interactive login)

## Installation

1. Install Playwright:
   ```bash
   npm install -g playwright
   playwright install chromium
   ```

2. Build the adapter and login tool:
   ```bash
   cd chatgpt-adapter-main
   make build
   ```

## Usage

### Interactive Login

The adapter now supports interactive login. When no cookies are provided, it will automatically:

1. Open a browser window
2. Navigate to Claude.ai
3. Prompt you to log in manually
4. Once you type 'Y' after logging in, it will save the cookies for future use

To manually trigger the interactive login process:

```bash
./bin/login -service claude
```

This will create a `claude_cookies.json` file in the current directory.

### API Endpoints

The adapter provides the following models:

- `web_claude`: Basic Claude model
- `web_claude/chat`: Chat completion model

### Authentication

Authentication is done using cookies from Claude.ai. You can provide these cookies in the `Authorization` header of your API requests.

Example:
```
Authorization: Bearer {"sessionKey":"your-session-key","sessionId":"your-session-id"}
```

If no cookies are provided, the adapter will automatically launch an interactive login process.

### Example Request

```json
{
  "model": "web_claude/chat",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant."
    },
    {
      "role": "user",
      "content": "Hello, how are you today?"
    }
  ],
  "stream": true
}
```

## Configuration

The adapter supports the following configuration options:

- `web_claude.debug`: Enable debug mode (default: false)

## Building as Executable

To build the adapter as a standalone executable:

```bash
cd chatgpt-adapter-main
make build
```

This will create executables in the `bin/` directory:
- `chatgpt-adapter`: The main adapter executable
- `login`: The interactive login tool

To install these executables to your system:

```bash
make install
```

## Limitations

- Image inputs are not currently supported
- Performance may be slower than using the official API

## Troubleshooting

If you encounter issues:

1. Try the interactive login process to refresh your cookies
2. Check that Playwright is properly installed
3. Enable debug mode to see detailed logs
4. Verify that you can manually log in to Claude.ai in your browser
