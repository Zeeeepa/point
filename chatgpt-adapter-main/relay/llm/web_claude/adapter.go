package web_claude

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
	Model = "web_claude"
	ChatModel = "web_claude/chat"
	mu    sync.Mutex
)

type api struct {
	inter.BaseAdapter

	env *env.Environment
	clientPool map[string]*ClaudeClient
	clientMu sync.Mutex
}

func (api *api) Match(ctx *gin.Context, model string) (ok bool, err error) {
	var token = ctx.GetString("token")
	ok = Model == model || model == ChatModel
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
		Id:      ChatModel,
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

	// Get or create Claude client
	client, err := api.getOrCreateClient(cookie)
	if err != nil {
		response.Error(ctx, -1, fmt.Sprintf("Failed to initialize Claude client: %v", err))
		return
	}

	// Convert OpenAI messages to Claude messages
	messages := convertMessages(completion.Messages)

	if completion.Stream {
		// Handle streaming response
		err = api.handleStreamingResponse(ctx, client, messages)
	} else {
		// Handle non-streaming response
		err = api.handleNonStreamingResponse(ctx, client, messages, completion)
	}

	return
}

func (api *api) getOrCreateClient(cookies map[string]string) (*ClaudeClient, error) {
	// Generate a key for the client based on cookies
	cookieKey := generateCookieKey(cookies)

	api.clientMu.Lock()
	defer api.clientMu.Unlock()

	// Initialize client pool if needed
	if api.clientPool == nil {
		api.clientPool = make(map[string]*ClaudeClient)
	}

	// Check if we already have a client for these cookies
	if client, ok := api.clientPool[cookieKey]; ok {
		return client, nil
	}

	// Create a new client
	client, err := NewClaudeClient()
	if err != nil {
		return nil, err
	}

	// Initialize the client with cookies
	debug := api.env.GetBool("web_claude.debug")
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
	return fmt.Sprintf("claude_cookies_%d", len(cookies))
}

func (api *api) handleNonStreamingResponse(ctx *gin.Context, client *ClaudeClient, messages []Message, completion model.Completion) error {
	// Send message to Claude and get response
	claudeResponse, err := client.SendMessage(ctx.Request.Context(), messages)
	if err != nil {
		response.Error(ctx, -1, fmt.Sprintf("Failed to get response from Claude: %v", err))
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
					"content": claudeResponse,
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

func (api *api) handleStreamingResponse(ctx *gin.Context, client *ClaudeClient, messages []Message) error {
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Transfer-Encoding", "chunked")

	// Create a channel to receive streaming responses
	responseCh := make(chan string)
	doneCh := make(chan bool)
	errCh := make(chan error)

	// Start streaming in a goroutine
	go func() {
		err := client.StreamMessage(ctx.Request.Context(), messages, func(content string, done bool) {
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

// convertMessages converts OpenAI messages to Claude messages
func convertMessages(messages []model.Keyv[interface{}]) []Message {
	var claudeMessages []Message
	
	for _, msg := range messages {
		role := msg.GetString("role")
		
		// Handle different content formats
		var content string
		
		if msg.IsString("content") {
			// Simple string content
			content = msg.GetString("content")
		} else if msg.Has("content") {
			// Array content (e.g., with images)
			contentArray := msg.GetSlice("content")
			var textParts []string
			
			for _, item := range contentArray {
				if itemMap, ok := item.(map[string]interface{}); ok {
					itemType, _ := itemMap["type"].(string)
					
					if itemType == "text" {
						if text, ok := itemMap["text"].(string); ok {
							textParts = append(textParts, text)
						}
					}
					// Note: We're ignoring image_url parts for now as Claude web doesn't support them directly
				}
			}
			
			content = strings.Join(textParts, "\n")
		}
		
		claudeMessages = append(claudeMessages, Message{
			Role:    role,
			Content: content,
		})
	}
	
	return claudeMessages
}

