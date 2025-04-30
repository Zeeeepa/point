package web_copilot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// CopilotClient is a client for interacting with GitHub Copilot's web interface
type CopilotClient struct {
	browserManager *BrowserManager
	logger         *log.Logger
	mu             sync.Mutex
}

// NewCopilotClient creates a new Copilot client
func NewCopilotClient() (*CopilotClient, error) {
	// Setup logging
	logFile, err := os.OpenFile("copilot_client.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	logger := log.New(logFile, "COPILOT_CLIENT: ", log.LstdFlags|log.Lshortfile)
	logger.Println("Initializing new Copilot client")

	return &CopilotClient{
		logger: logger,
	}, nil
}

// Initialize initializes the Copilot client with cookies
func (c *CopilotClient) Initialize(cookies map[string]string, debug bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Println("Initializing Copilot client")

	// Check if cookies are empty and prompt for interactive login
	if len(cookies) == 0 {
		c.logger.Println("No cookies provided, initiating interactive login")
		if err := c.runInteractiveLogin(); err != nil {
			return fmt.Errorf("interactive login failed: %v", err)
		}
		
		// Load cookies from file
		cookieFile := "github_cookies.json"
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
	browserManager, err := NewBrowserManager("~/.copilot-browser", false, debug)
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

	// Navigate to Copilot editor
	if err := browserManager.NavigateToCopilotEditor(); err != nil {
		browserManager.Close()
		return fmt.Errorf("failed to navigate to Copilot editor: %v", err)
	}

	c.browserManager = browserManager
	return nil
}

// runInteractiveLogin launches the login executable to perform interactive login
func (c *CopilotClient) runInteractiveLogin() error {
	c.logger.Println("Running interactive login process")
	
	// Find the login executable
	execPath, err := findLoginExecutable()
	if err != nil {
		return fmt.Errorf("failed to find login executable: %v", err)
	}
	
	// Run the login process
	cmd := exec.Command(execPath, "-service", "github")
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
func (c *CopilotClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.browserManager != nil {
		c.browserManager.Close()
		c.browserManager = nil
	}
}

// ParseCookieString parses a cookie string into a map
func (c *CopilotClient) ParseCookieString(cookieStr string) map[string]string {
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
func (c *CopilotClient) ParseCookieJSON(jsonStr string) (map[string]string, error) {
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

// GetCompletion sends code context to Copilot and returns the suggestion
func (c *CopilotClient) GetCompletion(ctx context.Context, codeContext string, language string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.browserManager == nil {
		return "", errors.New("Copilot client not initialized")
	}

	// Send the code context to Copilot
	suggestion, err := c.browserManager.SendCodeContext(ctx, codeContext, language)
	if err != nil {
		return "", fmt.Errorf("failed to get suggestion from Copilot: %v", err)
	}

	return suggestion, nil
}

// StreamCompletion sends code context to Copilot and streams the suggestion
func (c *CopilotClient) StreamCompletion(ctx context.Context, codeContext string, language string, callback func(string, bool)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.browserManager == nil {
		return errors.New("Copilot client not initialized")
	}

	// Stream the suggestion
	err := c.browserManager.StreamSuggestion(ctx, codeContext, language, callback)
	if err != nil {
		return fmt.Errorf("failed to stream suggestion from Copilot: %v", err)
	}

	return nil
}

// DetectLanguage attempts to detect the programming language from the code context
func (c *CopilotClient) DetectLanguage(codeContext string) string {
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

// SendCodeCompletion sends a code context to Copilot and returns the suggested completion
func (c *CopilotClient) SendCodeCompletion(ctx context.Context, codeContext string, language string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.browserManager == nil {
		return "", errors.New("client not initialized")
	}

	// Send the code context to Copilot
	suggestion, err := c.browserManager.SendCodeContext(ctx, codeContext, language)
	if err != nil {
		return "", fmt.Errorf("failed to send code context to Copilot: %v", err)
	}

	return suggestion, nil
}

// StreamCodeCompletion sends a code context to Copilot and streams the suggested completion
func (c *CopilotClient) StreamCodeCompletion(ctx context.Context, codeContext string, language string, callback func(string, bool)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.browserManager == nil {
		return errors.New("client not initialized")
	}

	// Stream the suggestion
	err := c.browserManager.StreamSuggestion(ctx, codeContext, language, callback)
	if err != nil {
		return fmt.Errorf("failed to stream suggestion from Copilot: %v", err)
	}

	return nil
}
