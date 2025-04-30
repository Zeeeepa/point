# ChatGPT Adapter with Web AI Integration

This project provides OpenAI API compatibility for various AI services, including web-based services like Claude AI and GitHub Copilot through browser automation.

## Features

- OpenAI API compatibility for multiple AI services
- Browser automation for web-based AI services
- Interactive login process for web services
- Streaming responses
- Conversation history support

## Web AI Adapters

This project includes adapters for web-based AI services:

- **Web Claude**: Interact with Claude AI through browser automation
- **Web Copilot**: Use GitHub Copilot through browser automation

## Installation

### Prerequisites

- Go 1.18 or higher
- Playwright for browser automation

### Install Playwright

```bash
npm install -g playwright
playwright install chromium
```

### Build the Project

```bash
# Clone the repository
git clone https://github.com/Zeeeepa/point.git
cd point/chatgpt-adapter-main

# Build the project
make build
```

This will create executables in the `bin/` directory:
- `chatgpt-adapter`: The main adapter executable
- `login`: The interactive login tool

### Install to System Path

```bash
make install
```

## Interactive Login Process

The web adapters support an interactive login process that:

1. Opens a browser window
2. Navigates to the service login page
3. Prompts you to log in manually
4. Saves the cookies for future use once you confirm

### Manual Login

To manually trigger the interactive login process:

```bash
# For Claude AI
./bin/login -service claude

# For GitHub Copilot
./bin/login -service github
```

### Automatic Login

The adapters will automatically launch the interactive login process when no cookies are provided in the API request.

## Usage

### Start the Server

```bash
./bin/chatgpt-adapter
```

### API Endpoints

The server provides OpenAI API compatible endpoints:

- `POST /v1/chat/completions` - For chat completions
- `POST /v1/completions` - For text completions

### Example Requests

#### Claude AI

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "web_claude/chat",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Hello, how are you today?"}
    ],
    "stream": true
  }'
```

#### GitHub Copilot

```bash
curl http://localhost:8080/v1/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "web_copilot/github",
    "prompt": "function calculateTotal(items) {\n  // Calculate the total price of all items\n  ",
    "stream": true
  }'
```

## Configuration

Configuration options can be set in the `config.yaml` file or through environment variables.

### Web Claude Options

- `web_claude.debug`: Enable debug mode (default: false)

### Web Copilot Options

- `web_copilot.debug`: Enable debug mode (default: false)

## Troubleshooting

If you encounter issues:

1. Try the interactive login process to refresh your cookies
2. Check that Playwright is properly installed
3. Enable debug mode to see detailed logs
4. Verify that you can manually log in to the service in your browser

## License

This project is licensed under the MIT License - see the LICENSE file for details.

