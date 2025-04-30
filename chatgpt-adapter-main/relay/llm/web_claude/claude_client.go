package web_claude

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
	"path/filepath"
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

	// Check if cookies are empty and prompt for interactive login
	if len(cookies) == 0 {
		c.logger.Println("No cookies provided, initiating interactive login")
		if err := c.runInteractiveLogin(); err != nil {
			return fmt.Errorf("interactive login failed: %v", err)
		}
		
		// Load cookies from file
		cookieFile := "claude_cookies.json"
		cookieData, err := os.ReadFile(cookieFile)
		if err != nil {
			return fmt.Errorf("failed to read cookie file after login: %v", err)
		}
		
		if err := json.Unmarshal(cookieData, &cookies); err != nil {
			return fmt.Errorf("failed to parse cookie file: %v", err)
		}
		
		c.logger.Printf("Loaded %d cookies from interactive login", len(cookies))
	}

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

// runInteractiveLogin launches the login executable to perform interactive login
func (c *ClaudeClient) runInteractiveLogin() error {
	c.logger.Println("Running interactive login process")
	
	// Find the login executable
	execPath, err := findLoginExecutable()
	if err != nil {
		return fmt.Errorf("failed to find login executable: %v", err)
	}
	
	// Run the login process
	cmd := exec.Command(execPath, "-service", "claude")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("login process failed: %v", err)
	}
	
	return nil
}

// findLoginExecutable locates the login executable
func findLoginExecutable() (string, error) {
	// Try to find the executable in the same directory as the current executable
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get current executable path: %v", err)
	}
	
	execDir := strings.TrimSuffix(execPath, filepath.Base(execPath))
	loginPath := filepath.Join(execDir, "login")
	
	// Check if the login executable exists
	if _, err := os.Stat(loginPath); err == nil {
		return loginPath, nil
	}
	
	// Try to find it in the PATH
	loginPath, err = exec.LookPath("login")
	if err == nil {
		return loginPath, nil
	}
	
	// Try to build it if it doesn't exist
	goPath, err := exec.LookPath("go")
	if err != nil {
		return "", fmt.Errorf("could not find Go executable: %v", err)
	}
	
	// Determine the path to the login source
	srcPath := filepath.Join(execDir, "..", "cmd", "login")
	if _, err := os.Stat(srcPath); err != nil {
		return "", fmt.Errorf("login source not found at %s: %v", srcPath, err)
	}
	
	// Build the login executable
	buildCmd := exec.Command(goPath, "build", "-o", loginPath, srcPath)
	if err := buildCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to build login executable: %v", err)
	}
	
	return loginPath, nil
}

// Close closes the client and releases resources
func (c *ClaudeClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.browserManager != nil {
		c.browserManager.Close()
		c.browserManager = nil
	}
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
		return "", errors.New("client not initialized")
	}

	// Format the prompt from messages
	prompt := formatPrompt(messages)

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
		return errors.New("client not initialized")
	}

	// Format the prompt from messages
	prompt := formatPrompt(messages)

	// Stream the response
	err := c.browserManager.StreamResponse(ctx, prompt, callback)
	if err != nil {
		return fmt.Errorf("failed to stream message from Claude: %v", err)
	}

	return nil
}

// formatPrompt formats a list of messages into a single prompt string
func formatPrompt(messages []Message) string {
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
