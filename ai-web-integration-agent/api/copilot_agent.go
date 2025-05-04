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

// CopilotAgent handles interactions with GitHub Copilot
type CopilotAgent struct {
	config Config
	logger *log.Logger
	session *Session
	mu     sync.Mutex
}

// NewCopilotAgent creates a new Copilot agent
func NewCopilotAgent(config Config) (*CopilotAgent, error) {
	// Setup logging
	logFile, err := os.OpenFile("copilot_agent.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	logger := log.New(logFile, "COPILOT_AGENT: ", log.LstdFlags|log.Lshortfile)
	logger.Println("Initializing Copilot agent")

	// Create a new session
	session, err := NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	// Login to GitHub
	if err := session.LoginToGitHub(); err != nil {
		session.Close()
		return nil, fmt.Errorf("GitHub login failed: %v", err)
	}

	return &CopilotAgent{
		config:  config,
		logger:  logger,
		session: session,
	}, nil
}

// ProcessChatCompletion processes a chat completion request for Copilot
func (a *CopilotAgent) ProcessChatCompletion(ctx context.Context, req ChatCompletionRequest) (interface{}, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.logger.Println("Processing chat completion request for Copilot")

	// Extract code context and language from the messages
	codeContext, language := extractCodeContext(req.Messages)

	if req.Stream {
		// Handle streaming response
		return a.handleStreamingChatCompletion(ctx, req, codeContext, language)
	}

	// Handle non-streaming response
	return a.handleNonStreamingChatCompletion(ctx, req, codeContext, language)
}

// handleNonStreamingChatCompletion handles a non-streaming chat completion request
func (a *CopilotAgent) handleNonStreamingChatCompletion(ctx context.Context, req ChatCompletionRequest, codeContext string, language string) (interface{}, error) {
	// Send the code context to Copilot
	suggestion, err := a.session.UseGitHubCopilot(codeContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestion from Copilot: %v", err)
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
					Content: suggestion,
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
func (a *CopilotAgent) handleStreamingChatCompletion(ctx context.Context, req ChatCompletionRequest, codeContext string, language string) (interface{}, error) {
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

		// Send the code context to Copilot and stream the suggestion
		var lastContent string
		var fullSuggestion string

		// Create a context with timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		// Start monitoring Copilot's suggestion
		suggestionCh := make(chan string)
		doneCh := make(chan bool)
		errCh := make(chan error)

		go func() {
			// Navigate to Copilot and prepare the editor
			if err := a.session.NavigateToCopilotEditor(); err != nil {
				errCh <- err
				return
			}

			// Send the code context to Copilot
			if err := a.session.SendCodeContextToCopilot(codeContext, language); err != nil {
				errCh <- err
				return
			}

			// Monitor Copilot's suggestion
			for {
				select {
				case <-timeoutCtx.Done():
					doneCh <- true
					return
				default:
					// Extract current suggestion
					currentSuggestion, err := a.session.ExtractCopilotSuggestion()
					if err != nil {
						continue
					}

					// If suggestion has changed, send the update
					if currentSuggestion != fullSuggestion {
						suggestionCh <- currentSuggestion
						fullSuggestion = currentSuggestion
					}

					// Check if Copilot is still generating
					isGenerating, err := a.session.IsCopilotGenerating()
					if err != nil {
						continue
					}

					if !isGenerating {
						// Double-check after a short delay
						time.Sleep(1 * time.Second)
						isGenerating, _ = a.session.IsCopilotGenerating()
						if !isGenerating {
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
			case content := <-suggestionCh:
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

// ProcessCompletion processes a completion request for Copilot
func (a *CopilotAgent) ProcessCompletion(ctx context.Context, req CompletionRequest) (interface{}, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.logger.Println("Processing completion request for Copilot")

	// Extract language from the prompt
	codeContext, language := extractLanguageFromPrompt(req.Prompt)

	if req.Stream {
		// Handle streaming response
		return a.handleStreamingCompletion(ctx, req, codeContext, language)
	}

	// Handle non-streaming response
	return a.handleNonStreamingCompletion(ctx, req, codeContext, language)
}

// handleNonStreamingCompletion handles a non-streaming completion request
func (a *CopilotAgent) handleNonStreamingCompletion(ctx context.Context, req CompletionRequest, codeContext string, language string) (interface{}, error) {
	// Send the code context to Copilot
	suggestion, err := a.session.UseGitHubCopilot(codeContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestion from Copilot: %v", err)
	}

	// Format the response
	return CompletionResponse{
		ID:      fmt.Sprintf("cmpl-%d", time.Now().Unix()),
		Object:  "text_completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []struct {
			Text         string `json:"text"`
			Index        int    `json:"index"`
			FinishReason string `json:"finish_reason"`
		}{
			{
				Text:         suggestion,
				Index:        0,
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

// handleStreamingCompletion handles a streaming completion request
func (a *CopilotAgent) handleStreamingCompletion(ctx context.Context, req CompletionRequest, codeContext string, language string) (interface{}, error) {
	// Create a channel for streaming responses
	streamCh := make(chan StreamChunk)

	// Start streaming in a goroutine
	go func() {
		defer close(streamCh)

		// Send the code context to Copilot and stream the suggestion
		var lastContent string
		var fullSuggestion string

		// Create a context with timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		// Start monitoring Copilot's suggestion
		suggestionCh := make(chan string)
		doneCh := make(chan bool)
		errCh := make(chan error)

		go func() {
			// Navigate to Copilot and prepare the editor
			if err := a.session.NavigateToCopilotEditor(); err != nil {
				errCh <- err
				return
			}

			// Send the code context to Copilot
			if err := a.session.SendCodeContextToCopilot(codeContext, language); err != nil {
				errCh <- err
				return
			}

			// Monitor Copilot's suggestion
			for {
				select {
				case <-timeoutCtx.Done():
					doneCh <- true
					return
				default:
					// Extract current suggestion
					currentSuggestion, err := a.session.ExtractCopilotSuggestion()
					if err != nil {
						continue
					}

					// If suggestion has changed, send the update
					if currentSuggestion != fullSuggestion {
						suggestionCh <- currentSuggestion
						fullSuggestion = currentSuggestion
					}

					// Check if Copilot is still generating
					isGenerating, err := a.session.IsCopilotGenerating()
					if err != nil {
						continue
					}

					if !isGenerating {
						// Double-check after a short delay
						time.Sleep(1 * time.Second)
						isGenerating, _ = a.session.IsCopilotGenerating()
						if !isGenerating {
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
			case content := <-suggestionCh:
				// Only send the new content since the last update
				newContent := strings.TrimPrefix(content, lastContent)
				lastContent = content

				// Create response chunk
				chunk := CompletionStreamResponse{
					ID:      fmt.Sprintf("cmpl-%d", time.Now().Unix()),
					Object:  "text_completion",
					Created: time.Now().Unix(),
					Model:   req.Model,
					Choices: []struct {
						Text         string  `json:"text"`
						Index        int     `json:"index"`
						FinishReason *string `json:"finish_reason"`
					}{
						{
							Text:         newContent,
							Index:        0,
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
				finalChunk := CompletionStreamResponse{
					ID:      fmt.Sprintf("cmpl-%d", time.Now().Unix()),
					Object:  "text_completion",
					Created: time.Now().Unix(),
					Model:   req.Model,
					Choices: []struct {
						Text         string  `json:"text"`
						Index        int     `json:"index"`
						FinishReason *string `json:"finish_reason"`
					}{
						{
							Text:         "",
							Index:        0,
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

// extractCodeContext extracts code context and language from chat messages
func extractCodeContext(messages []ChatCompletionMessage) (string, string) {
	var codeContext string
	var language string

	// Use the last user message as the code context
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			codeContext = messages[i].Content
			break
		}
	}

	// Check if language is specified in the code context
	languagePrefix := "language:"
	for _, line := range strings.Split(codeContext, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), languagePrefix) {
			language = strings.TrimSpace(line[len(languagePrefix):])
			// Remove the language line from the code context
			codeContext = strings.Replace(codeContext, line, "", 1)
			break
		}
	}

	// If language is not specified, try to detect it
	if language == "" {
		language = detectLanguage(codeContext)
	}

	return codeContext, language
}

// extractLanguageFromPrompt extracts code context and language from a prompt
func extractLanguageFromPrompt(prompt string) (string, string) {
	var language string

	// Check if language is specified in the prompt
	languagePrefix := "language:"
	for _, line := range strings.Split(prompt, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), languagePrefix) {
			language = strings.TrimSpace(line[len(languagePrefix):])
			// Remove the language line from the prompt
			prompt = strings.Replace(prompt, line, "", 1)
			break
		}
	}

	// If language is not specified, try to detect it
	if language == "" {
		language = detectLanguage(prompt)
	}

	return prompt, language
}

// detectLanguage attempts to detect the programming language from the code context
func detectLanguage(codeContext string) string {
	// Simple language detection based on file extensions in the code context
	if strings.Contains(codeContext, ".py") || strings.HasPrefix(codeContext, "import ") || strings.HasPrefix(codeContext, "from ") || strings.Contains(codeContext, "def ") {
		return "python"
	}
	if strings.Contains(codeContext, ".js") || strings.Contains(codeContext, "function ") || strings.Contains(codeContext, "const ") || strings.Contains(codeContext, "let ") {
		return "javascript"
	}
	if strings.Contains(codeContext, ".ts") || strings.Contains(codeContext, "interface ") || strings.Contains(codeContext, ": string") || strings.Contains(codeContext, ": number") {
		return "typescript"
	}
	if strings.Contains(codeContext, ".java") || strings.Contains(codeContext, "public class ") || strings.Contains(codeContext, "private ") || strings.Contains(codeContext, "protected ") {
		return "java"
	}
	if strings.Contains(codeContext, ".go") || strings.Contains(codeContext, "package ") || strings.Contains(codeContext, "func ") || strings.Contains(codeContext, "import (") {
		return "go"
	}
	if strings.Contains(codeContext, ".rb") || strings.Contains(codeContext, "def ") || strings.Contains(codeContext, "require ") || strings.Contains(codeContext, "class ") && strings.Contains(codeContext, "end") {
		return "ruby"
	}
	if strings.Contains(codeContext, ".php") || strings.Contains(codeContext, "<?php") || strings.Contains(codeContext, "namespace ") || strings.Contains(codeContext, "function ") && strings.Contains(codeContext, "$") {
		return "php"
	}
	if strings.Contains(codeContext, ".cs") || strings.Contains(codeContext, "using ") || strings.Contains(codeContext, "namespace ") || strings.Contains(codeContext, "class ") && strings.Contains(codeContext, "{") {
		return "csharp"
	}
	if strings.Contains(codeContext, ".cpp") || strings.Contains(codeContext, ".hpp") || strings.Contains(codeContext, "#include <") || strings.Contains(codeContext, "std::") {
		return "cpp"
	}
	if strings.Contains(codeContext, ".c") || strings.Contains(codeContext, "#include <") || strings.Contains(codeContext, "int main(") {
		return "c"
	}
	if strings.Contains(codeContext, ".rs") || strings.Contains(codeContext, "fn ") || strings.Contains(codeContext, "let mut ") || strings.Contains(codeContext, "impl ") {
		return "rust"
	}
	if strings.Contains(codeContext, ".html") || strings.Contains(codeContext, "<!DOCTYPE") || strings.Contains(codeContext, "<html") {
		return "html"
	}
	if strings.Contains(codeContext, ".css") || strings.Contains(codeContext, "{") && strings.Contains(codeContext, "}") && strings.Contains(codeContext, ":") {
		return "css"
	}
	if strings.Contains(codeContext, ".sh") || strings.Contains(codeContext, "#!/bin/bash") || strings.Contains(codeContext, "#!/bin/sh") {
		return "shell"
	}
	if strings.Contains(codeContext, ".sql") || strings.Contains(codeContext, "SELECT ") || strings.Contains(codeContext, "INSERT INTO ") || strings.Contains(codeContext, "CREATE TABLE ") {
		return "sql"
	}
	
	// Default to JavaScript if we can't detect the language
	return "javascript"
}

