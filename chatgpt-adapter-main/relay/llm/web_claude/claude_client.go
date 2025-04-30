package web_claude

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// ClaudeClient is a client for interacting with Claude's web interface
type ClaudeClient struct {
	browserManager *BrowserManager
	logger         *log.Logger
	mu             sync.Mutex
}

// NewClaudeClient creates a new Claude client
func NewClaudeClient() (*ClaudeClient, error) {
	// Setup logging
	logFile, err := os.OpenFile("claude_client.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	logger := log.New(logFile, "CLAUDE_CLIENT: ", log.LstdFlags|log.Lshortfile)
	logger.Println("Initializing new Claude client")

	return &ClaudeClient{
		logger: logger,
	}, nil
}

// Initialize initializes the Claude client with cookies
func (c *ClaudeClient) Initialize(cookies map[string]string, debug bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Println("Initializing Claude client")

	// Create browser manager
	browserManager, err := NewBrowserManager("~/.claude-browser", false, debug)
	if err != nil {
		return fmt.Errorf("failed to create browser manager: %v", err)
	}

	// Initialize browser
	if err := browserManager.Initialize(); err != nil {
		browserManager.Close()
		return fmt.Errorf("failed to initialize browser: %v", err)
	}

	// Load cookies
	if err := browserManager.LoadCookies(cookies); err != nil {
		browserManager.Close()
		return fmt.Errorf("failed to load cookies: %v", err)
	}

	// Navigate to Claude chat
	if err := browserManager.NavigateToClaudeChat(); err != nil {
		browserManager.Close()
		return fmt.Errorf("failed to navigate to Claude chat: %v", err)
	}

	c.browserManager = browserManager
	return nil
}

// Close closes the Claude client
func (c *ClaudeClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.browserManager != nil {
		c.browserManager.Close()
		c.browserManager = nil
	}
	c.logger.Println("Claude client closed")
}

// ParseCookieString parses a cookie string into a map
func (c *ClaudeClient) ParseCookieString(cookieStr string) map[string]string {
	cookies := make(map[string]string)
	
	// Split the cookie string by semicolons
	parts := strings.Split(cookieStr, ";")
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		// Split each part by the first equals sign
		if idx := strings.Index(part, "="); idx > 0 {
			name := strings.TrimSpace(part[:idx])
			value := strings.TrimSpace(part[idx+1:])
			cookies[name] = value
		}
	}
	
	return cookies
}

// ParseCookieJSON parses a JSON string of cookies into a map
func (c *ClaudeClient) ParseCookieJSON(jsonStr string) (map[string]string, error) {
	cookies := make(map[string]string)
	
	// Try to parse as an array of cookie objects
	var cookieArray []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	
	err := json.Unmarshal([]byte(jsonStr), &cookieArray)
	if err == nil {
		for _, cookie := range cookieArray {
			cookies[cookie.Name] = cookie.Value
		}
		return cookies, nil
	}
	
	// Try to parse as a map
	err = json.Unmarshal([]byte(jsonStr), &cookies)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cookie JSON: %v", err)
	}
	
	return cookies, nil
}

// SendMessage sends a message to Claude and returns the response
func (c *ClaudeClient) SendMessage(ctx context.Context, messages []Message) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.browserManager == nil {
		return "", errors.New("Claude client not initialized")
	}

	// Format messages into a prompt
	prompt := c.formatMessages(messages)

	// Send the prompt to Claude
	response, err := c.browserManager.SendMessage(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to send message to Claude: %v", err)
	}

	return response, nil
}

// StreamMessage sends a message to Claude and streams the response
func (c *ClaudeClient) StreamMessage(ctx context.Context, messages []Message, callback func(string, bool)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.browserManager == nil {
		return errors.New("Claude client not initialized")
	}

	// Format messages into a prompt
	prompt := c.formatMessages(messages)

	// Stream the response
	err := c.browserManager.StreamResponse(ctx, prompt, callback)
	if err != nil {
		return fmt.Errorf("failed to stream response from Claude: %v", err)
	}

	return nil
}

// formatMessages formats a slice of messages into a prompt for Claude
func (c *ClaudeClient) formatMessages(messages []Message) string {
	var prompt strings.Builder

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			prompt.WriteString("System: ")
			prompt.WriteString(msg.Content)
			prompt.WriteString("\n\n")
		case "user":
			prompt.WriteString("Human: ")
			prompt.WriteString(msg.Content)
			prompt.WriteString("\n\n")
		case "assistant":
			prompt.WriteString("Assistant: ")
			prompt.WriteString(msg.Content)
			prompt.WriteString("\n\n")
		}
	}

	// Add the final assistant prefix to prompt Claude to respond
	prompt.WriteString("Assistant: ")

	return prompt.String()
}

// Message represents a message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ConversationContext represents a conversation context
type ConversationContext struct {
	Messages []Message
}

// NewConversationContext creates a new conversation context
func NewConversationContext() *ConversationContext {
	return &ConversationContext{
		Messages: []Message{},
	}
}

// AddMessage adds a message to the conversation context
func (cc *ConversationContext) AddMessage(role, content string) {
	cc.Messages = append(cc.Messages, Message{
		Role:    role,
		Content: content,
	})
}

// GetMessages returns all messages in the conversation context
func (cc *ConversationContext) GetMessages() []Message {
	return cc.Messages
}

