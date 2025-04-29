# Point: Technical Architecture

This document outlines the technical architecture of the Point project, including component interactions, data flow, API specifications, and technical decisions.

## System Architecture Overview

![System Architecture Diagram](https://via.placeholder.com/800x600?text=System+Architecture+Diagram)

The Point system consists of the following major components:

1. **User Interface (UI)**
   - React-based frontend application
   - Communicates with Backend API

2. **Backend API**
   - Node.js/Express application
   - Manages configuration
   - Coordinates component interactions

3. **ChatGPT Adapter**
   - Go-based server
   - Adapts various AI services to OpenAI API format
   - Handles request routing

4. **AI Gateway**
   - Rust-based API gateway
   - Manages provider connections
   - Handles load balancing and failover

5. **Web Integration Agent**
   - Go/Python hybrid application
   - Manages browser automation
   - Integrates with web-based AI interfaces

6. **Configuration Store**
   - Database for storing configurations
   - Supports multiple configuration profiles

7. **Monitoring System**
   - Collects metrics and logs
   - Provides visualization and alerting

## Component Interactions

### 1. User Interface to Backend API

The UI communicates with the Backend API using RESTful HTTP requests:

```
UI → HTTP Request → Backend API → HTTP Response → UI
```

Key interactions:
- Configuration management
- Provider testing
- Status monitoring
- Log retrieval

### 2. Backend API to ChatGPT Adapter

The Backend API communicates with the ChatGPT Adapter for configuration and status:

```
Backend API → Configuration Updates → ChatGPT Adapter
ChatGPT Adapter → Status Updates → Backend API
```

Key interactions:
- Provider configuration
- Endpoint mapping
- Service status monitoring

### 3. Backend API to AI Gateway

The Backend API communicates with the AI Gateway for provider management:

```
Backend API → Provider Configuration → AI Gateway
AI Gateway → Metrics and Status → Backend API
```

Key interactions:
- Provider API key management
- Routing rule configuration
- Performance metrics collection

### 4. Backend API to Web Integration Agent

The Backend API communicates with the Web Integration Agent for browser automation:

```
Backend API → Browser Configuration → Web Integration Agent
Web Integration Agent → Session Status → Backend API
```

Key interactions:
- Browser profile management
- Authentication configuration
- Session monitoring

### 5. Client Applications to Point System

Client applications communicate with the Point system through the standard OpenAI API format:

```
Client App → OpenAI API Request → Point System → AI Service → Response → Client App
```

Key interactions:
- Chat completions
- Embeddings
- Image generation
- Model listing

## Data Flow

### 1. Configuration Flow

```
UI → Configuration Update → Backend API → Configuration Store → Component Configuration → Components
```

1. User updates configuration in UI
2. UI sends configuration to Backend API
3. Backend API validates and stores configuration
4. Backend API updates relevant components
5. Components apply new configuration

### 2. Request Flow

```
Client → API Request → ChatGPT Adapter → Provider Selection → AI Gateway/Web Agent → AI Service → Response → Client
```

1. Client sends request to Point system
2. ChatGPT Adapter receives and processes request
3. Provider is selected based on routing rules
4. Request is forwarded to appropriate provider via AI Gateway or Web Agent
5. Response is received from AI service
6. Response is formatted and returned to client

### 3. Monitoring Flow

```
Components → Metrics/Logs → Backend API → Processing → UI Display
```

1. Components generate metrics and logs
2. Backend API collects and processes data
3. UI requests and displays monitoring information

## API Specifications

### 1. Backend API

#### Provider Management

```
GET /api/providers
```
- Returns list of all configured providers
- Query parameters:
  - `type`: Filter by provider type (api, web)
  - `status`: Filter by status (active, inactive)
  - `page`: Page number
  - `limit`: Items per page

```
GET /api/providers/:id
```
- Returns details of a specific provider
- Path parameters:
  - `id`: Provider ID

```
POST /api/providers
```
- Creates a new provider
- Request body:
  - `type`: Provider type (api, web)
  - `name`: Custom name
  - `provider`: Provider identifier
  - `config`: Provider-specific configuration

```
PUT /api/providers/:id
```
- Updates an existing provider
- Path parameters:
  - `id`: Provider ID
- Request body:
  - `name`: Custom name
  - `config`: Provider-specific configuration

```
DELETE /api/providers/:id
```
- Deletes a provider
- Path parameters:
  - `id`: Provider ID

```
POST /api/providers/:id/test
```
- Tests a provider connection
- Path parameters:
  - `id`: Provider ID
- Request body:
  - `prompt`: Test prompt
  - `model`: Model to test
  - `parameters`: Additional parameters

#### Endpoint Configuration

```
GET /api/endpoints
```
- Returns list of all endpoints
- Query parameters:
  - `status`: Filter by status (active, inactive)

```
GET /api/endpoints/:id
```
- Returns details of a specific endpoint
- Path parameters:
  - `id`: Endpoint ID

```
PUT /api/endpoints/:id
```
- Updates an endpoint configuration
- Path parameters:
  - `id`: Endpoint ID
- Request body:
  - `active`: Boolean to enable/disable
  - `mapping`: Provider mapping configuration

```
POST /api/endpoints/:id/test
```
- Tests an endpoint configuration
- Path parameters:
  - `id`: Endpoint ID
- Request body:
  - `request`: Sample request

#### Configuration Management

```
GET /api/configs
```
- Returns list of all saved configurations
- Query parameters:
  - `page`: Page number
  - `limit`: Items per page

```
GET /api/configs/:id
```
- Returns details of a specific configuration
- Path parameters:
  - `id`: Configuration ID

```
POST /api/configs
```
- Saves current configuration
- Request body:
  - `name`: Configuration name
  - `description`: Configuration description

```
PUT /api/configs/:id
```
- Updates a saved configuration
- Path parameters:
  - `id`: Configuration ID
- Request body:
  - `name`: Configuration name
  - `description`: Configuration description

```
DELETE /api/configs/:id
```
- Deletes a saved configuration
- Path parameters:
  - `id`: Configuration ID

```
POST /api/configs/:id/activate
```
- Activates a saved configuration
- Path parameters:
  - `id`: Configuration ID

```
GET /api/configs/export
```
- Exports current configuration as JSON

```
POST /api/configs/import
```
- Imports configuration from JSON
- Request body:
  - `config`: Configuration JSON

#### Monitoring

```
GET /api/requests
```
- Returns list of recent requests
- Query parameters:
  - `provider`: Filter by provider
  - `endpoint`: Filter by endpoint
  - `status`: Filter by status
  - `from`: Start timestamp
  - `to`: End timestamp
  - `page`: Page number
  - `limit`: Items per page

```
GET /api/requests/:id
```
- Returns details of a specific request
- Path parameters:
  - `id`: Request ID

```
GET /api/requests/stats
```
- Returns request statistics
- Query parameters:
  - `provider`: Filter by provider
  - `endpoint`: Filter by endpoint
  - `from`: Start timestamp
  - `to`: End timestamp
  - `groupBy`: Group by field (hour, day, week, month)

```
GET /api/status
```
- Returns system status

```
GET /api/status/health
```
- Returns health check status

```
GET /api/logs
```
- Returns system logs
- Query parameters:
  - `level`: Filter by log level
  - `component`: Filter by component
  - `from`: Start timestamp
  - `to`: End timestamp
  - `search`: Search term
  - `page`: Page number
  - `limit`: Items per page

### 2. OpenAI-Compatible API

The Point system exposes the following OpenAI-compatible endpoints:

```
POST /v1/chat/completions
```
- Chat completions endpoint
- Request body follows OpenAI format
- Additional parameters:
  - `provider`: Override provider selection
  - `fallback_providers`: Array of fallback providers

```
POST /v1/completions
```
- Completions endpoint
- Request body follows OpenAI format
- Additional parameters:
  - `provider`: Override provider selection
  - `fallback_providers`: Array of fallback providers

```
POST /v1/embeddings
```
- Embeddings endpoint
- Request body follows OpenAI format
- Additional parameters:
  - `provider`: Override provider selection

```
POST /v1/images/generations
```
- Image generation endpoint
- Request body follows OpenAI format
- Additional parameters:
  - `provider`: Override provider selection

```
GET /v1/models
```
- Models listing endpoint
- Query parameters:
  - `provider`: Filter by provider

## Database Schema

### Providers Table

```sql
CREATE TABLE providers (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  type VARCHAR(50) NOT NULL,
  provider VARCHAR(100) NOT NULL,
  config JSONB NOT NULL,
  status VARCHAR(50) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);
```

### Endpoints Table

```sql
CREATE TABLE endpoints (
  id UUID PRIMARY KEY,
  path VARCHAR(255) NOT NULL,
  active BOOLEAN NOT NULL DEFAULT true,
  mapping JSONB NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);
```

### Configurations Table

```sql
CREATE TABLE configurations (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  is_default BOOLEAN NOT NULL DEFAULT false,
  config JSONB NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);
```

### Requests Table

```sql
CREATE TABLE requests (
  id UUID PRIMARY KEY,
  timestamp TIMESTAMP NOT NULL,
  endpoint VARCHAR(255) NOT NULL,
  provider_id UUID,
  model VARCHAR(255),
  status INTEGER NOT NULL,
  response_time INTEGER NOT NULL,
  tokens_input INTEGER,
  tokens_output INTEGER,
  cost DECIMAL(10, 6),
  request JSONB NOT NULL,
  response JSONB,
  FOREIGN KEY (provider_id) REFERENCES providers(id)
);
```

### Logs Table

```sql
CREATE TABLE logs (
  id UUID PRIMARY KEY,
  timestamp TIMESTAMP NOT NULL,
  level VARCHAR(20) NOT NULL,
  component VARCHAR(100) NOT NULL,
  message TEXT NOT NULL,
  context JSONB
);
```

## Technical Decisions

### 1. Frontend Framework: React with TypeScript

**Decision**: Use React with TypeScript for the frontend.

**Rationale**:
- Strong typing helps prevent runtime errors
- Component-based architecture for reusability
- Large ecosystem and community support
- Good performance for interactive UIs
- TypeScript provides better IDE support and documentation

### 2. Backend Framework: Node.js with Express

**Decision**: Use Node.js with Express for the backend API.

**Rationale**:
- JavaScript/TypeScript consistency with frontend
- Non-blocking I/O for handling multiple concurrent requests
- Rich ecosystem of libraries
- Easy integration with various databases
- Good performance for API services

### 3. Database: PostgreSQL

**Decision**: Use PostgreSQL for the primary database.

**Rationale**:
- JSONB support for flexible configuration storage
- Strong reliability and data integrity
- Good performance for mixed read/write workloads
- Advanced query capabilities
- Open-source with strong community support

### 4. Authentication: JWT

**Decision**: Use JWT (JSON Web Tokens) for authentication.

**Rationale**:
- Stateless authentication
- Reduced database load
- Scalability across multiple servers
- Support for fine-grained permissions
- Wide library support

### 5. API Design: RESTful with OpenAPI

**Decision**: Use RESTful API design with OpenAPI specification.

**Rationale**:
- Familiar pattern for developers
- Clear resource-oriented structure
- Good caching support
- OpenAPI provides documentation and client generation
- Easy to test and debug

### 6. Containerization: Docker

**Decision**: Use Docker for containerization.

**Rationale**:
- Consistent environments across development and production
- Isolation of components
- Simplified deployment
- Scalability
- Good orchestration options (Docker Compose, Kubernetes)

## Security Considerations

### 1. API Key Management

- API keys stored encrypted in the database
- Keys never logged or exposed in responses
- Option to use environment variables instead of database storage
- Regular key rotation encouraged

### 2. Authentication and Authorization

- JWT-based authentication
- Role-based access control
- Token expiration and refresh mechanism
- HTTPS required for all communications

### 3. Input Validation

- All API inputs validated against schemas
- Sanitization of user inputs
- Rate limiting to prevent abuse
- Request size limits

### 4. Dependency Security

- Regular security audits of dependencies
- Automated vulnerability scanning
- Minimal dependency footprint
- Pinned dependency versions

### 5. Logging and Monitoring

- Security-relevant events logged
- Anomaly detection
- Failed authentication attempts monitored
- Regular log review process

## Deployment Architecture

### Development Environment

```
Local Machine → Docker Compose → All Components
```

- Docker Compose for local development
- SQLite for development database
- Hot reloading for frontend and backend
- Local mocks for external services

### Production Environment

#### Option 1: Single Server

```
Server → Docker Compose → All Components → PostgreSQL
```

- All components on a single server
- Docker Compose for orchestration
- PostgreSQL for database
- Nginx for reverse proxy and SSL termination

#### Option 2: Distributed

```
Load Balancer → Frontend Servers → API Servers → Component Servers → Database Cluster
```

- Multiple servers for scalability
- Kubernetes for orchestration
- PostgreSQL cluster for database
- Redis for caching
- Separate servers for each component

## Performance Considerations

### 1. Caching

- Response caching where appropriate
- Configuration caching
- Model list caching
- Redis for distributed caching

### 2. Database Optimization

- Indexing for common queries
- Connection pooling
- Query optimization
- Regular maintenance

### 3. Frontend Optimization

- Code splitting
- Lazy loading
- Bundle size optimization
- CDN for static assets

### 4. API Gateway Optimization

- Request batching
- Connection reuse
- Efficient routing algorithms
- Response streaming

## Monitoring and Observability

### 1. Metrics Collection

- Request count and latency
- Error rates
- Resource utilization
- Provider-specific metrics

### 2. Logging

- Structured logging
- Log levels (debug, info, warn, error)
- Context-rich logs
- Log aggregation

### 3. Alerting

- Error rate thresholds
- Latency thresholds
- Resource utilization thresholds
- Provider availability

### 4. Dashboards

- System overview
- Provider performance
- Request patterns
- Error analysis

## Conclusion

This technical architecture provides a comprehensive blueprint for implementing the Point system. It outlines the component interactions, data flow, API specifications, and technical decisions that form the foundation of the system. As development progresses, this document should be updated to reflect any changes or refinements to the architecture.

