package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// ClaudeAgent handles interactions with Claude
type ClaudeAgent struct {
	config Config
	logger *log.Logger
	session *Session
	mu     sync.Mutex
}

// NewClaudeAgent creates a new Claude agent
func NewClaudeAgent(config Config) (*ClaudeAgent, error) {
	// Setup logging
	logFile, err := os.OpenFile("claude_agent.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	logger := log.New(logFile, "CLAUDE_AGENT: ", log.LstdFlags|log.Lshortfile)
	logger.Println("Initializing Claude agent")

	// Create a new session
	session, err := NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	// Login to Claude
	if err := session.LoginToClaude(); err != nil {
		session.Close()
		return nil, fmt.Errorf("Claude login failed: %v", err)
	}

	return &ClaudeAgent{
		config:  config,
		logger:  logger,
		session: session,
	}, nil
}

// ProcessChatCompletion processes a chat completion request for Claude
func (a *ClaudeAgent) ProcessChatCompletion(ctx context.Context, req ChatCompletionRequest) (interface{}, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.logger.Println("Processing chat completion request for Claude")

	// Format the prompt from messages
	prompt := formatClaudePrompt(req.Messages)

	if req.Stream {
		// Handle streaming response
		return a.handleStreamingChatCompletion(ctx, req, prompt)
	}

	// Handle non-streaming response
	return a.handleNonStreamingChatCompletion(ctx, req, prompt)
}

// handleNonStreamingChatCompletion handles a non-streaming chat completion request
func (a *ClaudeAgent) handleNonStreamingChatCompletion(ctx context.Context, req ChatCompletionRequest, prompt string) (interface{}, error) {
	// Send the prompt to Claude
	response, err := a.session.AskClaude(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get response from Claude: %v", err)
	}

	// Format the response
	return ChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		}{
			{
				Index: 0,
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: response,
				},
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     0, // We don't have token counts
			CompletionTokens: 0,
			TotalTokens:      0,
		},
	}, nil
}

// handleStreamingChatCompletion handles a streaming chat completion request
func (a *ClaudeAgent) handleStreamingChatCompletion(ctx context.Context, req ChatCompletionRequest, prompt string) (interface{}, error) {
	// Create a channel for streaming responses
	streamCh := make(chan StreamChunk)

	// Start streaming in a goroutine
	go func() {
		defer close(streamCh)

		// Send initial response
		initialResponse := ChatCompletionStreamResponse{
			ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   req.Model,
			Choices: []struct {
				Index        int               `json:"index"`
				Delta        map[string]string `json:"delta"`
				FinishReason *string           `json:"finish_reason"`
			}{
				{
					Index: 0,
					Delta: map[string]string{
						"role": "assistant",
					},
					FinishReason: nil,
				},
			},
		}

		streamCh <- StreamChunk{
			Data: initialResponse,
			Done: false,
		}

		// Send the prompt to Claude and stream the response
		var lastContent string
		var fullResponse string

		// Create a context with timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()

		// Start monitoring Claude's response
		responseCh := make(chan string)
		doneCh := make(chan bool)
		errCh := make(chan error)

		go func() {
			// Send the prompt to Claude
			if err := a.session.SendPromptToClaude(prompt); err != nil {
				errCh <- err
				return
			}

			// Monitor Claude's response
			for {
				select {
				case <-timeoutCtx.Done():
					doneCh <- true
					return
				default:
					// Extract current response
					currentResponse, err := a.session.ExtractClaudeResponse()
					if err != nil {
						continue
					}

					// If response has changed, send the update
					if currentResponse != fullResponse {
						responseCh <- currentResponse
						fullResponse = currentResponse
					}

					// Check if Claude is still typing
					isTyping, err := a.session.IsClaudeTyping()
					if err != nil {
						continue
					}

					if !isTyping {
						// Double-check after a short delay
						time.Sleep(1 * time.Second)
						isTyping, _ = a.session.IsClaudeTyping()
						if !isTyping {
							doneCh <- true
							return
						}
					}

					time.Sleep(500 * time.Millisecond)
				}
			}
		}()

		// Process streaming responses
		for {
			select {
			case content := <-responseCh:
				// Only send the new content since the last update
				newContent := strings.TrimPrefix(content, lastContent)
				lastContent = content

				// Create response chunk
				chunk := ChatCompletionStreamResponse{
					ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),
					Model:   req.Model,
					Choices: []struct {
						Index        int               `json:"index"`
						Delta        map[string]string `json:"delta"`
						FinishReason *string           `json:"finish_reason"`
					}{
						{
							Index: 0,
							Delta: map[string]string{
								"content": newContent,
							},
							FinishReason: nil,
						},
					},
				}

				streamCh <- StreamChunk{
					Data: chunk,
					Done: false,
				}

			case <-doneCh:
				// Send final chunk
				finishReason := "stop"
				finalChunk := ChatCompletionStreamResponse{
					ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
					Object:  "chat.completion.chunk",
					Created: time.Now().Unix(),
					Model:   req.Model,
					Choices: []struct {
						Index        int               `json:"index"`
						Delta        map[string]string `json:"delta"`
						FinishReason *string           `json:"finish_reason"`
					}{
						{
							Index:        0,
							Delta:        map[string]string{},
							FinishReason: &finishReason,
						},
					},
				}

				streamCh <- StreamChunk{
					Data: finalChunk,
					Done: true,
				}
				return

			case err := <-errCh:
				a.logger.Printf("Error streaming response: %v", err)
				return

			case <-ctx.Done():
				// Client disconnected
				return
			}
		}
	}()

	return streamCh, nil
}

// ProcessCompletion processes a completion request for Claude
func (a *ClaudeAgent) ProcessCompletion(ctx context.Context, req CompletionRequest) (interface{}, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.logger.Println("Processing completion request for Claude")

	// Convert completion request to chat completion format
	chatReq := ChatCompletionRequest{
		Model: req.Model,
		Messages: []ChatCompletionMessage{
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      req.Stream,
	}

	// Process as chat completion
	result, err := a.ProcessChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	// If streaming, return as is
	if req.Stream {
		return result, nil
	}

	// Convert chat completion response to completion response
	chatResp := result.(ChatCompletionResponse)
	return CompletionResponse{
		ID:      chatResp.ID,
		Object:  "text_completion",
		Created: chatResp.Created,
		Model:   chatResp.Model,
		Choices: []struct {
			Text         string `json:"text"`
			Index        int    `json:"index"`
			FinishReason string `json:"finish_reason"`
		}{
			{
				Text:         chatResp.Choices[0].Message.Content,
				Index:        chatResp.Choices[0].Index,
				FinishReason: chatResp.Choices[0].FinishReason,
			},
		},
		Usage: chatResp.Usage,
	}, nil
}

// formatClaudePrompt formats a list of messages into a single prompt string for Claude
func formatClaudePrompt(messages []ChatCompletionMessage) string {
	var systemPrompt string
	var userPrompts []string

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			systemPrompt = msg.Content
		case "user":
			userPrompts = append(userPrompts, msg.Content)
		case "assistant":
			// For now, we ignore assistant messages as Claude web doesn't support them directly
			// In a more advanced implementation, we could simulate a conversation
		}
	}

	var prompt string
	if systemPrompt != "" {
		prompt = fmt.Sprintf("System: %s\n\n", systemPrompt)
	}

	if len(userPrompts) > 0 {
		prompt += strings.Join(userPrompts, "\n\n")
	}

	return prompt
}

