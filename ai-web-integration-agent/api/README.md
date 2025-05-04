# AI Web Integration API Server

This API server provides OpenAI API-compatible endpoints for Claude and GitHub Copilot web interfaces. It uses browser automation to interact with these services and expose them through a standardized API.

## Features

- OpenAI API-compatible endpoints for Claude and GitHub Copilot
- Support for both streaming and non-streaming responses
- Automatic session management and authentication
- Docker support for easy deployment

## Available Models

- `web_claude` - Claude web interface
- `web_claude/chat` - Claude chat interface
- `web_copilot` - GitHub Copilot web interface
- `web_copilot/github` - GitHub Copilot with GitHub context

## Prerequisites

- Go 1.18 or higher
- Chrome or Chromium browser
- Node.js and npm (for Playwright)

## Installation

### From Source

1. Clone the repository:
   ```bash
   git clone https://github.com/Zeeeepa/point.git
   cd point/ai-web-integration-agent
   ```

2. Build the API server:
   ```bash
   cd api
   go build -o web-integration-api .
   ```

3. Run the API server:
   ```bash
   ./web-integration-api
   ```

### Using Docker

1. Build the Docker image:
   ```bash
   docker build -t web-integration-api -f api/Dockerfile .
   ```

2. Run the Docker container:
   ```bash
   docker run -p 8080:8080 web-integration-api
   ```

### Using Docker Compose

1. Run the Docker Compose setup:
   ```bash
   docker-compose up -d
   ```

## Configuration

The API server can be configured using a JSON configuration file or command-line flags.

### Configuration File

Create a `config.json` file with the following structure:

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

### Command-Line Flags

- `--config`: Path to configuration file (default: `config.json`)
- `--port`: Port to listen on (default: `8080`)
- `--host`: Host to listen on (default: `0.0.0.0`)
- `--headless`: Run in headless mode (default: `false`)
- `--debug`: Enable debug mode (default: `false`)

## Authentication

The API server requires authentication with Claude and GitHub Copilot. On first run, it will open browser windows for you to log in to these services. After logging in, the sessions will be saved for future use.

## API Endpoints

### Chat Completions

```
POST /v1/chat/completions
```

Request:
```json
{
  "model": "web_claude",
  "messages": [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Hello, how are you?"}
  ],
  "temperature": 0.7,
  "stream": false
}
```

### Completions

```
POST /v1/completions
```

Request:
```json
{
  "model": "web_copilot",
  "prompt": "function calculateSum(a, b) {",
  "temperature": 0.7,
  "stream": false
}
```

### Models

```
GET /v1/models
```

### Health Check

```
GET /health
```

## Integration with chatgpt-adapter

This API server is designed to work with the chatgpt-adapter framework. The web_claude and web_copilot adapters in the chatgpt-adapter-main repository can be configured to use this API server as a backend.

## License

MIT

