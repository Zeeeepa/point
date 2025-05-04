# Integration Guide for chatgpt-adapter-main

This guide explains how to integrate the AI Web Integration API Server with the chatgpt-adapter-main framework.

## Overview

The AI Web Integration API Server provides OpenAI API-compatible endpoints for Claude and GitHub Copilot web interfaces. The chatgpt-adapter-main framework already includes adapters for these services, but they use direct browser automation. This integration allows the adapters to use the API server instead, which provides better stability and resource management.

## Prerequisites

1. The AI Web Integration API Server running on a host accessible from the chatgpt-adapter-main server
2. Access to the chatgpt-adapter-main codebase

## Configuration

### 1. Update the web_claude adapter

Edit the `chatgpt-adapter-main/relay/llm/web_claude/adapter.go` file to use the API server:

```go
// Add a new field to the api struct
type api struct {
    inter.BaseAdapter

    env *env.Environment
    clientPool map[string]*ClaudeClient
    clientMu sync.Mutex
    apiBaseURL string // New field for API server URL
}

// Update the NewAdapter function in ctor.go
func NewAdapter(env *env.Environment) inter.Adapter {
    apiBaseURL := env.GetString("web_claude.api_base_url")
    if apiBaseURL == "" {
        apiBaseURL = "http://localhost:8080/v1" // Default API server URL
    }
    
    return &api{
        env: env,
        apiBaseURL: apiBaseURL,
    }
}

// Update the Completion function to use the API server
func (api *api) Completion(ctx *gin.Context) (err error) {
    var (
        cookie, _  = common.GetGinValue[map[string]string](ctx, "token")
        completion = common.GetGinCompletion(ctx)
    )

    // If API server URL is set, use it
    if api.apiBaseURL != "" {
        return api.handleAPIServerCompletion(ctx, completion)
    }

    // Otherwise, use the existing implementation
    // ... existing code ...
}

// Add a new function to handle API server completion
func (api *api) handleAPIServerCompletion(ctx *gin.Context, completion model.Completion) error {
    // Create HTTP client
    client := &http.Client{
        Timeout: 120 * time.Second,
    }

    // Create request body
    reqBody, err := json.Marshal(completion)
    if err != nil {
        response.Error(ctx, -1, fmt.Sprintf("Failed to marshal request: %v", err))
        return err
    }

    // Create request
    req, err := http.NewRequest("POST", api.apiBaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
    if err != nil {
        response.Error(ctx, -1, fmt.Sprintf("Failed to create request: %v", err))
        return err
    }

    // Set headers
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Accept", "application/json")

    // Handle streaming
    if completion.Stream {
        return api.handleAPIServerStreamingCompletion(ctx, client, req)
    }

    // Handle non-streaming
    resp, err := client.Do(req)
    if err != nil {
        response.Error(ctx, -1, fmt.Sprintf("Failed to send request: %v", err))
        return err
    }
    defer resp.Body.Close()

    // Check response status
    if resp.StatusCode != http.StatusOK {
        body, _ := ioutil.ReadAll(resp.Body)
        response.Error(ctx, -1, fmt.Sprintf("API server error: %s", string(body)))
        return fmt.Errorf("API server error: %s", string(body))
    }

    // Copy response to client
    ctx.Header("Content-Type", "application/json")
    _, err = io.Copy(ctx.Writer, resp.Body)
    return err
}

// Add a new function to handle API server streaming completion
func (api *api) handleAPIServerStreamingCompletion(ctx *gin.Context, client *http.Client, req *http.Request) error {
    // Set streaming headers
    ctx.Header("Content-Type", "text/event-stream")
    ctx.Header("Cache-Control", "no-cache")
    ctx.Header("Connection", "keep-alive")
    ctx.Header("Transfer-Encoding", "chunked")

    // Send request
    resp, err := client.Do(req)
    if err != nil {
        response.Error(ctx, -1, fmt.Sprintf("Failed to send request: %v", err))
        return err
    }
    defer resp.Body.Close()

    // Check response status
    if resp.StatusCode != http.StatusOK {
        body, _ := ioutil.ReadAll(resp.Body)
        response.Error(ctx, -1, fmt.Sprintf("API server error: %s", string(body)))
        return fmt.Errorf("API server error: %s", string(body))
    }

    // Stream response to client
    reader := bufio.NewReader(resp.Body)
    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                break
            }
            return err
        }

        // Write line to client
        _, err = ctx.Writer.WriteString(line)
        if err != nil {
            return err
        }
        ctx.Writer.Flush()

        // Check if client disconnected
        if ctx.Request.Context().Err() != nil {
            return ctx.Request.Context().Err()
        }
    }

    return nil
}
```

### 2. Update the web_copilot adapter

Make similar changes to the `chatgpt-adapter-main/relay/llm/web_copilot/adapter.go` file.

### 3. Update the configuration

Add the API server URL to your configuration:

```yaml
web_claude:
  api_base_url: "http://localhost:8080/v1"

web_copilot:
  api_base_url: "http://localhost:8080/v1"
```

## Cookie Management

The API server handles cookie management internally, so you don't need to pass cookies from the chatgpt-adapter-main framework. However, you will need to log in to Claude and GitHub Copilot through the API server's browser interface on first use.

## Testing the Integration

1. Start the AI Web Integration API Server:
   ```bash
   cd ai-web-integration-agent/api
   go run main.go
   ```

2. Start the chatgpt-adapter-main server:
   ```bash
   cd chatgpt-adapter-main
   go run main.go
   ```

3. Send a test request to the chatgpt-adapter-main server:
   ```bash
   curl -X POST http://localhost:8080/v1/chat/completions \
     -H "Content-Type: application/json" \
     -d '{
       "model": "web_claude",
       "messages": [
         {"role": "user", "content": "Hello, how are you?"}
       ]
     }'
   ```

## Troubleshooting

- If you encounter connection errors, make sure the API server is running and accessible from the chatgpt-adapter-main server.
- If you encounter authentication errors, make sure you've logged in to Claude and GitHub Copilot through the API server's browser interface.
- If you encounter timeout errors, try increasing the timeout in the HTTP client.

## Advanced Configuration

You can configure the API server with additional options:

- `headless`: Set to `true` to run the browser in headless mode (no UI)
- `debug_mode`: Set to `true` to enable debug logging and screenshots
- `browser_user_data_dir`: Set to a custom directory to store browser data

See the API server's README.md for more configuration options.

