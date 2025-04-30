package web_copilot

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"chatgpt-adapter/core/common"
	"chatgpt-adapter/core/gin/inter"
	"chatgpt-adapter/core/gin/model"
	"chatgpt-adapter/core/gin/response"
	"github.com/gin-gonic/gin"
	"github.com/iocgo/sdk/env"
	"github.com/iocgo/sdk/stream"
)

var (
	Model = "web_copilot"
	GithubModel = "web_copilot/github"
	mu    sync.Mutex
)

type api struct {
	inter.BaseAdapter

	env *env.Environment
	clientPool map[string]*CopilotClient
	clientMu sync.Mutex
}

func (api *api) Match(ctx *gin.Context, model string) (ok bool, err error) {
	var token = ctx.GetString("token")
	ok = Model == model || model == GithubModel
	if ok {
		password := api.env.GetString("server.password")
		if password != "" && password != token {
			err = response.UnauthorizedError
			return
		}
	}
	return
}

func (*api) Models() (slice []model.Model) {
	slice = append(slice, model.Model{
		Id:      Model,
		Object:  "model",
		Created: 1686935002,
		By:      Model + "-adapter",
	})
	slice = append(slice, model.Model{
		Id:      GithubModel,
		Object:  "model",
		Created: 1686935002,
		By:      Model + "-adapter",
	})
	return
}

func (api *api) Completion(ctx *gin.Context) (err error) {
	var (
		cookie, _  = common.GetGinValue[map[string]string](ctx, "token")
		completion = common.GetGinCompletion(ctx)
	)

	// Get or create Copilot client
	client, err := api.getOrCreateClient(cookie)
	if err != nil {
		response.Error(ctx, -1, fmt.Sprintf("Failed to initialize Copilot client: %v", err))
		return
	}

	// Extract code context from the prompt
	codeContext, language := api.extractCodeContext(completion)

	if completion.Stream {
		// Handle streaming response
		err = api.handleStreamingResponse(ctx, client, codeContext, language)
	} else {
		// Handle non-streaming response
		err = api.handleNonStreamingResponse(ctx, client, codeContext, language, completion)
	}

	return
}

func (api *api) getOrCreateClient(cookies map[string]string) (*CopilotClient, error) {
	// Generate a key for the client based on cookies
	cookieKey := generateCookieKey(cookies)

	api.clientMu.Lock()
	defer api.clientMu.Unlock()

	// Initialize client pool if needed
	if api.clientPool == nil {
		api.clientPool = make(map[string]*CopilotClient)
	}

	// Check if we already have a client for these cookies
	if client, ok := api.clientPool[cookieKey]; ok {
		return client, nil
	}

	// Create a new client
	client, err := NewCopilotClient()
	if err != nil {
		return nil, err
	}

	// Initialize the client with cookies
	debug := api.env.GetBool("web_copilot.debug")
	if err := client.Initialize(cookies, debug); err != nil {
		client.Close()
		return nil, err
	}

	// Store the client in the pool
	api.clientPool[cookieKey] = client
	return client, nil
}

// generateCookieKey generates a unique key for a set of cookies
func generateCookieKey(cookies map[string]string) string {
	// Sort cookie names to ensure consistent key generation
	var names []string
	for name := range cookies {
		names = append(names, name)
	}
	
	// For simplicity, just use the number of cookies as a key
	// In a production environment, you might want a more robust solution
	return fmt.Sprintf("copilot_cookies_%d", len(cookies))
}

func (api *api) extractCodeContext(completion model.Completion) (codeContext string, language string) {
	// Extract code context from the prompt
	if len(completion.Messages) > 0 {
		lastMessage := completion.Messages[len(completion.Messages)-1]
		if lastMessage.Has("content") {
			content := lastMessage.GetString("content")
			
			// Check if language is specified in the prompt
			languagePrefix := "language:"
			for _, line := range strings.Split(content, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(strings.ToLower(line), languagePrefix) {
					language = strings.TrimSpace(line[len(languagePrefix):])
					// Remove the language line from the content
					content = strings.Replace(content, line, "", 1)
					break
				}
			}
			
			codeContext = content
		}
	}
	
	return
}

func (api *api) handleNonStreamingResponse(ctx *gin.Context, client *CopilotClient, codeContext string, language string, completion model.Completion) error {
	// If language is not specified, try to detect it
	if language == "" {
		language = client.DetectLanguage(codeContext)
	}
	
	// Send code context to Copilot and get suggestion
	suggestion, err := client.GetCompletion(ctx.Request.Context(), codeContext, language)
	if err != nil {
		response.Error(ctx, -1, fmt.Sprintf("Failed to get suggestion from Copilot: %v", err))
		return err
	}

	// Format response in OpenAI format
	resp := model.CompletionResponse{
		Id:      common.GenId(),
		Object:  "chat.completion",
		Created: common.GenTimestamp(),
		Model:   completion.Model,
		Choices: []model.Choice{
			{
				Index: 0,
				Message: map[string]interface{}{
					"role":    "assistant",
					"content": suggestion,
				},
				FinishReason: "stop",
			},
		},
		Usage: model.Usage{
			PromptTokens:     0,
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}

	response.JSON(ctx, resp)
	return nil
}

func (api *api) handleStreamingResponse(ctx *gin.Context, client *CopilotClient, codeContext string, language string) error {
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Transfer-Encoding", "chunked")

	// If language is not specified, try to detect it
	if language == "" {
		language = client.DetectLanguage(codeContext)
	}

	// Create a channel to receive streaming responses
	responseCh := make(chan string)
	doneCh := make(chan bool)
	errCh := make(chan error)

	// Start streaming in a goroutine
	go func() {
		err := client.StreamCompletion(ctx.Request.Context(), codeContext, language, func(content string, done bool) {
			if done {
				doneCh <- true
				return
			}
			responseCh <- content
		})
		if err != nil {
			errCh <- err
		}
	}()

	// Send initial response
	initialResponse := model.CompletionStreamResponse{
		Id:      common.GenId(),
		Object:  "chat.completion.chunk",
		Created: common.GenTimestamp(),
		Model:   Model,
		Choices: []model.StreamChoice{
			{
				Index: 0,
				Delta: map[string]interface{}{
					"role": "assistant",
				},
				FinishReason: nil,
			},
		},
	}
	
	initialJSON, _ := json.Marshal(initialResponse)
	ctx.Writer.Write([]byte("data: " + string(initialJSON) + "\n\n"))
	ctx.Writer.Flush()

	var lastContent string
	
	// Process streaming responses
	for {
		select {
		case content := <-responseCh:
			// Only send the new content since the last update
			newContent := strings.TrimPrefix(content, lastContent)
			lastContent = content
			
			// Create response chunk
			chunk := model.CompletionStreamResponse{
				Id:      common.GenId(),
				Object:  "chat.completion.chunk",
				Created: common.GenTimestamp(),
				Model:   Model,
				Choices: []model.StreamChoice{
					{
						Index: 0,
						Delta: map[string]interface{}{
							"content": newContent,
						},
						FinishReason: nil,
					},
				},
			}
			
			// Send chunk
			chunkJSON, _ := json.Marshal(chunk)
			ctx.Writer.Write([]byte("data: " + string(chunkJSON) + "\n\n"))
			ctx.Writer.Flush()
			
		case <-doneCh:
			// Send final chunk
			finalChunk := model.CompletionStreamResponse{
				Id:      common.GenId(),
				Object:  "chat.completion.chunk",
				Created: common.GenTimestamp(),
				Model:   Model,
				Choices: []model.StreamChoice{
					{
						Index:        0,
						Delta:        map[string]interface{}{},
						FinishReason: "stop",
					},
				},
			}
			
			finalJSON, _ := json.Marshal(finalChunk)
			ctx.Writer.Write([]byte("data: " + string(finalJSON) + "\n\n"))
			ctx.Writer.Write([]byte("data: [DONE]\n\n"))
			ctx.Writer.Flush()
			return nil
			
		case err := <-errCh:
			// Send error
			response.Error(ctx, -1, fmt.Sprintf("Streaming error: %v", err))
			return err
			
		case <-ctx.Request.Context().Done():
			// Client disconnected
			return ctx.Request.Context().Err()
		}
	}
}

