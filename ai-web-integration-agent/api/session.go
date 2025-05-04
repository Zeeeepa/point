package api

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

// Session represents a browser session
type Session struct {
	ctx    context.Context
	cancel context.CancelFunc
	config Config
	logger *log.Logger
}

// NewSession creates a new browser session
func NewSession(config Config) (*Session, error) {
	// Setup logging
	logFile, err := os.OpenFile("session.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	logger := log.New(logFile, "SESSION: ", log.LstdFlags|log.Lshortfile)
	logger.Println("Initializing browser session")

	// Expand user directory in browser user data dir
	if config.BrowserUserDataDir == "~/.browser-agent" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %v", err)
		}
		config.BrowserUserDataDir = filepath.Join(homeDir, ".browser-agent")
	}

	// Create browser user data directory if it doesn't exist
	if err := os.MkdirAll(config.BrowserUserDataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create browser user data directory: %v", err)
	}

	// Create screenshots directory if it doesn't exist
	if err := os.MkdirAll(config.ScreenshotDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create screenshots directory: %v", err)
	}

	// Setup Chrome options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "IsolateOrigins,site-per-process"),
		chromedp.Flag("disable-site-isolation-trials", true),
		chromedp.UserDataDir(config.BrowserUserDataDir),
	)

	if config.Headless {
		opts = append(opts, chromedp.Headless)
	}

	// Create a new ExecAllocator
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Create a new browser context
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(logger.Printf))

	// Set a timeout for the browser context
	ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)

	return &Session{
		ctx:    ctx,
		cancel: cancel,
		config: config,
		logger: logger,
	}, nil
}

// Close closes the browser session
func (s *Session) Close() {
	s.logger.Println("Closing browser session")
	s.cancel()
}

// LoginToClaude logs in to Claude
func (s *Session) LoginToClaude() error {
	s.logger.Println("Logging in to Claude")

	// Navigate to Claude
	if err := chromedp.Run(s.ctx, chromedp.Navigate(s.config.ClaudeURL)); err != nil {
		return fmt.Errorf("failed to navigate to Claude: %v", err)
	}

	// Check if already logged in
	var isLoggedIn bool
	err := chromedp.Run(s.ctx, chromedp.Evaluate(`
		document.querySelector('textarea') !== null
	`, &isLoggedIn))

	if err != nil {
		return fmt.Errorf("failed to check if logged in to Claude: %v", err)
	}

	if isLoggedIn {
		s.logger.Println("Already logged in to Claude")
		return nil
	}

	// Take a screenshot
	if s.config.DebugMode {
		if err := s.takeScreenshot("claude_login.png"); err != nil {
			s.logger.Printf("Warning: Failed to take screenshot: %v", err)
		}
	}

	// Wait for user to log in manually
	s.logger.Println("Please log in to Claude manually in the opened browser window")
	fmt.Println("Please log in to Claude manually in the opened browser window")
	fmt.Println("Press Enter when you're done...")
	fmt.Scanln()

	// Check if login was successful
	err = chromedp.Run(s.ctx, chromedp.Evaluate(`
		document.querySelector('textarea') !== null
	`, &isLoggedIn))

	if err != nil {
		return fmt.Errorf("failed to check if logged in to Claude: %v", err)
	}

	if !isLoggedIn {
		return fmt.Errorf("failed to log in to Claude")
	}

	s.logger.Println("Successfully logged in to Claude")
	return nil
}

// LoginToGitHub logs in to GitHub
func (s *Session) LoginToGitHub() error {
	s.logger.Println("Logging in to GitHub")

	// Navigate to GitHub
	if err := chromedp.Run(s.ctx, chromedp.Navigate("https://github.com/login")); err != nil {
		return fmt.Errorf("failed to navigate to GitHub: %v", err)
	}

	// Check if already logged in
	var isLoggedIn bool
	err := chromedp.Run(s.ctx, chromedp.Evaluate(`
		document.querySelector('.logged-in') !== null
	`, &isLoggedIn))

	if err != nil {
		return fmt.Errorf("failed to check if logged in to GitHub: %v", err)
	}

	if isLoggedIn {
		s.logger.Println("Already logged in to GitHub")
		return nil
	}

	// Take a screenshot
	if s.config.DebugMode {
		if err := s.takeScreenshot("github_login.png"); err != nil {
			s.logger.Printf("Warning: Failed to take screenshot: %v", err)
		}
	}

	// Wait for user to log in manually
	s.logger.Println("Please log in to GitHub manually in the opened browser window")
	fmt.Println("Please log in to GitHub manually in the opened browser window")
	fmt.Println("Press Enter when you're done...")
	fmt.Scanln()

	// Check if login was successful
	err = chromedp.Run(s.ctx, chromedp.Evaluate(`
		document.querySelector('.logged-in') !== null
	`, &isLoggedIn))

	if err != nil {
		return fmt.Errorf("failed to check if logged in to GitHub: %v", err)
	}

	if !isLoggedIn {
		return fmt.Errorf("failed to log in to GitHub")
	}

	s.logger.Println("Successfully logged in to GitHub")
	return nil
}

// AskClaude sends a prompt to Claude and returns the response
func (s *Session) AskClaude(prompt string) (string, error) {
	s.logger.Println("Asking Claude")

	// Navigate to Claude
	if err := chromedp.Run(s.ctx, chromedp.Navigate(s.config.ClaudeURL)); err != nil {
		return "", fmt.Errorf("failed to navigate to Claude: %v", err)
	}

	// Wait for Claude to load
	if err := chromedp.Run(s.ctx, 
		chromedp.WaitVisible(`textarea`, chromedp.ByQuery),
	); err != nil {
		return "", fmt.Errorf("failed waiting for Claude input: %v", err)
	}

	// Clear existing text and type new prompt
	if err := chromedp.Run(s.ctx,
		chromedp.Click(`textarea`, chromedp.ByQuery),
		chromedp.KeyEvent("Control+a"), // Select all
		chromedp.KeyEvent("Delete"), // Delete selected
		chromedp.SendKeys(`textarea`, prompt, chromedp.ByQuery),
		chromedp.KeyEvent("Enter"), // Send the prompt
	); err != nil {
		return "", fmt.Errorf("failed to input prompt: %v", err)
	}

	// Wait for response to appear
	if err := s.WaitForClaudeToFinishTyping(s.ctx); err != nil {
		return "", fmt.Errorf("failed waiting for Claude to respond: %v", err)
	}

	// Extract the response
	var response string
	err := chromedp.Run(s.ctx, chromedp.Evaluate(`
		// Get all message containers
		const messages = document.querySelectorAll('div[role="article"]');
		// Get the latest message (Claude's response)
		const lastMessage = messages[messages.length - 1];
		return lastMessage ? lastMessage.innerText : "";
	`, &response))

	if err != nil {
		return "", fmt.Errorf("failed to extract Claude's response: %v", err)
	}

	return response, nil
}

// UseGitHubCopilot sends code to GitHub Copilot and returns the suggestion
func (s *Session) UseGitHubCopilot(code string) (string, error) {
	s.logger.Println("Using GitHub Copilot")

	// Navigate to GitHub Copilot
	if err := s.NavigateToCopilotEditor(); err != nil {
		return "", fmt.Errorf("failed to navigate to Copilot editor: %v", err)
	}

	// Detect language
	language := detectLanguage(code)

	// Send code to Copilot
	if err := s.SendCodeContextToCopilot(code, language); err != nil {
		return "", fmt.Errorf("failed to send code to Copilot: %v", err)
	}

	// Wait for suggestion to appear
	time.Sleep(3 * time.Second)

	// Extract the suggestion
	suggestion, err := s.ExtractCopilotSuggestion()
	if err != nil {
		return "", fmt.Errorf("failed to extract Copilot suggestion: %v", err)
	}

	return suggestion, nil
}

// takeScreenshot takes a screenshot of the current page
func (s *Session) takeScreenshot(filename string) error {
	var buf []byte
	if err := chromedp.Run(s.ctx, chromedp.FullScreenshot(&buf, 90)); err != nil {
		return fmt.Errorf("failed to take screenshot: %v", err)
	}

	if err := os.WriteFile(filepath.Join(s.config.ScreenshotDir, filename), buf, 0644); err != nil {
		return fmt.Errorf("failed to write screenshot: %v", err)
	}

	return nil
}

