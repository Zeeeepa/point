# Web Claude Adapter

This adapter provides OpenAI API compatibility for Claude AI through browser automation. It allows you to interact with Claude's web interface without using their official API.

## Features

- Browser automation to interact with Claude's web interface
- Conversation history support
- Streaming responses
- OpenAI API compatibility
- System, user, and assistant message support

## Requirements

- [Playwright](https://playwright.dev/) for browser automation
- Valid Claude.ai cookies for authentication

## Installation

1. Install Playwright:
   ```bash
   npm install -g playwright
   playwright install chromium
   ```

2. Extract your Claude.ai cookies using the provided script:
   ```bash
   python firefox-cookie-extractor.py -d claude.ai -f json -o claude_cookies.json
   ```

## Usage

### API Endpoints

The adapter provides the following models:

- `web_claude`: Basic Claude model
- `web_claude/chat`: Chat completion model

### Authentication

Authentication is done using cookies from Claude.ai. You need to provide these cookies in the `Authorization` header of your API requests.

Example:
```
Authorization: Bearer {"sessionKey":"your-session-key","sessionId":"your-session-id"}
```

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

## Limitations

- Image inputs are not currently supported
- The adapter requires a valid Claude.ai session
- Performance may be slower than using the official API

## Troubleshooting

If you encounter issues:

1. Make sure your Claude.ai cookies are valid and up-to-date
2. Check that Playwright is properly installed
3. Enable debug mode to see detailed logs
4. Verify that you can manually log in to Claude.ai in your browser

