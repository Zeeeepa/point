# Web Copilot Adapter

This adapter provides OpenAI API compatibility for GitHub Copilot through browser automation. It allows you to interact with GitHub Copilot's web interface without using their official API.

## Features

- Browser automation to interact with GitHub Copilot's web interface
- Code completion and generation
- Streaming responses
- OpenAI API compatibility
- Automatic language detection

## Requirements

- [Playwright](https://playwright.dev/) for browser automation
- Valid GitHub cookies with Copilot access

## Installation

1. Install Playwright:
   ```bash
   npm install -g playwright
   playwright install chromium
   ```

2. Extract your GitHub cookies using the provided script:
   ```bash
   python firefox-cookie-extractor.py -d github.com -f json -o github_cookies.json
   ```

## Usage

### API Endpoints

The adapter provides the following models:

- `web_copilot`: Basic Copilot model
- `web_copilot/github`: GitHub-specific Copilot model

### Authentication

Authentication is done using cookies from GitHub. You need to provide these cookies in the `Authorization` header of your API requests.

Example:
```
Authorization: Bearer {"user_session":"your-session-id","_gh_sess":"your-gh-sess"}
```

### Example Request

```json
{
  "model": "web_copilot/github",
  "messages": [
    {
      "role": "user",
      "content": "language: python\n\ndef calculate_fibonacci(n):\n    # Implement a function to calculate the nth Fibonacci number"
    }
  ],
  "stream": true
}
```

### Language Specification

You can specify the programming language in your prompt by including a line with `language:` followed by the language name. For example:

```
language: python
```

If not specified, the adapter will attempt to detect the language automatically based on the code context.

## Configuration

The adapter supports the following configuration options:

- `web_copilot.debug`: Enable debug mode (default: false)

## Limitations

- The adapter requires a valid GitHub session with Copilot access
- Performance may be slower than using the official API
- Some advanced Copilot features may not be available

## Troubleshooting

If you encounter issues:

1. Make sure your GitHub cookies are valid and up-to-date
2. Check that Playwright is properly installed
3. Enable debug mode to see detailed logs
4. Verify that you can manually log in to GitHub and access Copilot in your browser

