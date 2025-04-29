# Point: Universal AI API Adapter

Point is a comprehensive system for unifying access to various AI services through a standardized interface. It combines multiple components to provide seamless integration with both API-based and web-based AI services.

![Point Architecture](https://via.placeholder.com/800x400?text=Point+Architecture)

## Core Components

- **ChatGPT Adapter**: Go-based server that adapts various AI services to the OpenAI API format
- **AI Gateway**: Rust-based API gateway providing unified access to multiple LLM providers
- **AI Web Integration Agent**: Tool for integrating with web interfaces like Claude and GitHub Copilot
- **API Endpoints**: TypeScript implementations for various provider APIs
- **Custom Adapters**: Specialized adapters for specific services

## Features

- **Unified API Interface**: Access multiple AI services through a standardized OpenAI-compatible API
- **Web Interface Integration**: Connect to web-based AI interfaces without direct API access
- **Provider Management**: Configure, test, and manage multiple AI providers
- **Endpoint Redirection**: Control which endpoints are redirected to which providers
- **Configuration Management**: Save, edit, and share provider configurations

## User Interface Requirements

### 1. Provider Selection & Configuration

#### 1.1 API Provider Configuration
- **Requirement**: Allow users to configure API-based providers
- **Details**:
  - Display a comprehensive list of available API providers
  - For each provider, allow configuration of:
    - API key
    - Custom name
    - Available models (fetched from provider)
    - Base URL (for custom deployments)
    - OpenAI compatibility flag
  - Provide a "Test Connection" button to verify configuration
  - Allow saving, editing, and deleting provider configurations

#### 1.2 Web Interface Configuration
- **Requirement**: Allow users to configure web-based interfaces
- **Details**:
  - Support for web interfaces including:
    - Claude
    - GitHub Copilot
    - Cursor
    - Windsurf
    - DeepSeek (webchat)
    - Hugging Face Spaces
  - For each interface, allow configuration of:
    - Custom name
    - URL (for Hugging Face Spaces)
    - Browser automation settings
    - Session persistence options
  - Provide a "Test Connection" button to verify configuration
  - Allow saving, editing, and deleting interface configurations

### 2. Endpoint Redirection Configuration

#### 2.1 Endpoint Selection
- **Requirement**: Allow users to select which endpoint calls to redirect
- **Details**:
  - Display a list of standard OpenAI-compatible endpoints:
    - `/v1/chat/completions`
    - `/v1/completions`
    - `/v1/embeddings`
    - `/v1/images/generations`
    - `/v1/models`
  - Allow enabling/disabling redirection for each endpoint
  - Provide visual indicators for active redirections

#### 2.2 Provider Mapping
- **Requirement**: Allow mapping endpoints to specific providers
- **Details**:
  - For each endpoint, allow selection of:
    - Primary provider
    - Fallback providers (in order of preference)
    - Load balancing options
  - Support for model-specific routing (e.g., route `gpt-4` to OpenAI, `claude-3` to Anthropic)
  - Allow setting default providers for unmapped models

### 3. Configuration Management

#### 3.1 Configuration Profiles
- **Requirement**: Allow saving and loading configuration profiles
- **Details**:
  - Save complete configurations with custom names
  - Load saved configurations
  - Export configurations as JSON files
  - Import configurations from JSON files
  - Set a default configuration to load at startup

#### 3.2 Provider Cards
- **Requirement**: Provide a card-based interface for managing providers
- **Details**:
  - Display each configured provider as a card
  - Show key information on the card:
    - Provider name
    - Status (active/inactive)
    - Configured models
    - Last successful connection
  - Provide quick actions:
    - Enable/disable
    - Edit
    - Delete
    - Test connection

### 4. Monitoring and Logs

#### 4.1 Request Monitoring
- **Requirement**: Display information about API requests
- **Details**:
  - Show recent requests with:
    - Timestamp
    - Endpoint
    - Provider used
    - Status (success/failure)
    - Response time
  - Allow filtering by provider, endpoint, and status
  - Provide detailed view of request/response pairs

#### 4.2 Usage Statistics
- **Requirement**: Display usage statistics for providers
- **Details**:
  - Show usage metrics:
    - Requests per provider
    - Average response time
    - Error rate
    - Token usage (where applicable)
  - Display historical usage trends
  - Export statistics as CSV

## Technical Requirements

### 1. Frontend

#### 1.1 Technology Stack
- **Framework**: React with TypeScript
- **UI Library**: Material-UI or Tailwind CSS
- **State Management**: Redux or Context API
- **API Client**: Axios or Fetch API

#### 1.2 Responsive Design
- Support for desktop and tablet devices
- Minimum resolution: 1280x720
- Dark and light theme options

### 2. Backend Integration

#### 2.1 ChatGPT Adapter Integration
- REST API for configuration management
- WebSocket for real-time status updates
- Configuration persistence in local storage or database

#### 2.2 AI Gateway Integration
- REST API for provider configuration
- Configuration synchronization
- Health check endpoints

#### 2.3 Web Integration Agent
- API for browser automation configuration
- Session management
- Screenshot capture for debugging

### 3. Security

#### 3.1 API Key Management
- Secure storage of API keys
- Option to use environment variables
- Masking of sensitive information in UI and logs

#### 3.2 Authentication
- Basic authentication for UI access
- Role-based access control (optional)
- Session timeout and management

## Implementation Plan

### Phase 1: Core UI Framework (Weeks 1-2)

#### Week 1: Project Setup
- Set up React project with TypeScript
- Implement basic layout and navigation
- Create component library and style guide

#### Week 2: Provider Management UI
- Implement provider card components
- Create forms for API provider configuration
- Implement basic CRUD operations for providers

### Phase 2: Provider Configuration (Weeks 3-4)

#### Week 3: API Provider Implementation
- Implement API provider configuration screens
- Create provider testing functionality
- Develop model selection interface

#### Week 4: Web Interface Implementation
- Implement web interface configuration screens
- Create browser automation settings
- Develop session management for web interfaces

### Phase 3: Endpoint Configuration (Weeks 5-6)

#### Week 5: Endpoint Selection
- Implement endpoint selection interface
- Create visual indicators for active redirections
- Develop endpoint testing functionality

#### Week 6: Provider Mapping
- Implement provider mapping interface
- Create fallback configuration
- Develop model-specific routing

### Phase 4: Configuration Management (Weeks 7-8)

#### Week 7: Configuration Profiles
- Implement save/load functionality
- Create import/export features
- Develop default configuration setting

#### Week 8: Monitoring and Logs
- Implement request monitoring interface
- Create usage statistics dashboard
- Develop log viewing and filtering

### Phase 5: Integration and Testing (Weeks 9-10)

#### Week 9: Backend Integration
- Integrate with ChatGPT Adapter
- Integrate with AI Gateway
- Integrate with Web Integration Agent

#### Week 10: Testing and Refinement
- Perform comprehensive testing
- Fix bugs and issues
- Refine UI based on feedback

## Getting Started

### Prerequisites
- Node.js 16+
- Go 1.18+
- Rust 1.60+
- Chrome/Chromium browser (for web integration)

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/point.git
cd point

# Install dependencies
npm install

# Start the development server
npm start
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

