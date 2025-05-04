# AI Web Integration Agent

This repository contains tools for integrating with AI web interfaces like Claude and GitHub Copilot.

## Components

### 1. Browser Automation Tool

The main component is a browser automation tool that can:

- Log into Claude and GitHub Copilot
- Send prompts to Claude and get responses
- Send code to GitHub Copilot and get suggestions
- Integrate both services for a workflow where Claude provides guidance and Copilot generates code

### 2. API Server

The API server provides OpenAI API-compatible endpoints for Claude and GitHub Copilot web interfaces. It uses browser automation to interact with these services and expose them through a standardized API.

Features:
- OpenAI API-compatible endpoints for Claude and GitHub Copilot
- Support for both streaming and non-streaming responses
- Automatic session management and authentication
- Docker support for easy deployment

Available Models:
- `web_claude` - Claude web interface
- `web_claude/chat` - Claude chat interface
- `web_copilot` - GitHub Copilot web interface
- `web_copilot/github` - GitHub Copilot with GitHub context

## Getting Started

### API Server

See the [API Server README](api/README.md) for detailed instructions on how to set up and use the API server.

### Integration with chatgpt-adapter

The API server is designed to work with the chatgpt-adapter framework. See the [Integration Guide](api/INTEGRATION.md) for instructions on how to integrate the API server with the chatgpt-adapter-main framework.

## Prerequisites

- Go 1.18 or higher
- Chrome or Chromium browser
- Node.js and npm (for Playwright)

## License

MIT

