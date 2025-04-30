package web_claude

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
)

// BrowserManager handles browser automation for Claude web interface
type BrowserManager struct {
	pw         *playwright.Playwright
	browser    playwright.Browser
	context    playwright.BrowserContext
	page       playwright.Page
	userDataDir string
	logger     *log.Logger
	isHeadless bool
	debug      bool
}

// NewBrowserManager creates a new browser manager instance
func NewBrowserManager(userDataDir string, isHeadless bool, debug bool) (*BrowserManager, error) {
	// Setup logging
	logFile, err := os.OpenFile("claude_browser.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	logger := log.New(logFile, "CLAUDE_BROWSER: ", log.LstdFlags|log.Lshortfile)
	logger.Println("Initializing new browser manager")

	// Expand ~ to home directory if present
	if strings.HasPrefix(userDataDir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %v", err)
		}
		userDataDir = filepath.Join(home, userDataDir[1:])
	}

	// Create user data directory if it doesn't exist
	if err := os.MkdirAll(userDataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create user data directory: %v", err)
	}

	return &BrowserManager{
		userDataDir: userDataDir,
		logger:      logger,
		isHeadless:  isHeadless,
		debug:       debug,
	}, nil
}

// Initialize sets up the browser
func (bm *BrowserManager) Initialize() error {
	bm.logger.Println("Initializing browser")

	// Install Playwright if not already installed
	if err := bm.ensurePlaywrightInstalled(); err != nil {
		return err
	}

	// Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("could not start playwright: %v", err)
	}
	bm.pw = pw

	// Launch browser
	browserOpts := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(bm.isHeadless),
		Args: []string{
			"--disable-web-security",
			"--disable-features=IsolateOrigins,site-per-process",
		},
	}

	browser, err := bm.pw.Chromium.Launch(browserOpts)
	if err != nil {
		return fmt.Errorf("could not launch browser: %v", err)
	}
	bm.browser = browser

	// Create browser context with user data directory
	contextOpts := playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	}

	context, err := browser.NewContext(contextOpts)
	if err != nil {
		return fmt.Errorf("could not create browser context: %v", err)
	}
	bm.context = context

	// Create page
	page, err := context.NewPage()
	if err != nil {
		return fmt.Errorf("could not create page: %v", err)
	}
	bm.page = page

	// Set default timeout
	bm.page.SetDefaultTimeout(30000)

	return nil
}

// ensurePlaywrightInstalled makes sure Playwright is installed
func (bm *BrowserManager) ensurePlaywrightInstalled() error {
	// Check if playwright is installed by running a simple command
	cmd := exec.Command("playwright", "version")
	if err := cmd.Run(); err != nil {
		// If not installed, install it
		bm.logger.Println("Playwright not found, installing...")
		installCmd := exec.Command("npm", "install", "-g", "playwright")
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to install playwright: %v", err)
		}
		
		// Install browsers
		browserCmd := exec.Command("playwright", "install", "chromium")
		if err := browserCmd.Run(); err != nil {
			return fmt.Errorf("failed to install playwright browsers: %v", err)
		}
	}
	
	return nil
}

// Close closes the browser and cleans up resources
func (bm *BrowserManager) Close() {
	if bm.page != nil {
		bm.page.Close()
	}
	if bm.context != nil {
		bm.context.Close()
	}
	if bm.browser != nil {
		bm.browser.Close()
	}
	if bm.pw != nil {
		bm.pw.Stop()
	}
	bm.logger.Println("Browser resources closed")
}

// LoadCookies loads cookies from a map into the browser
func (bm *BrowserManager) LoadCookies(cookies map[string]string) error {
	bm.logger.Println("Loading cookies")
	
	// Convert cookies map to Playwright cookie format
	var playwrightCookies []playwright.OptionalCookie
	
	for name, value := range cookies {
		playwrightCookies = append(playwrightCookies, playwright.OptionalCookie{
			Name:   name,
			Value:  value,
			Domain: playwright.String(".claude.ai"),
			Path:   playwright.String("/"),
		})
	}
	
	// Set cookies in browser context
	if err := bm.context.AddCookies(playwrightCookies); err != nil {
		return fmt.Errorf("failed to set cookies: %v", err)
	}
	
	return nil
}

// NavigateToClaudeChat navigates to Claude chat interface
func (bm *BrowserManager) NavigateToClaudeChat() error {
	bm.logger.Println("Navigating to Claude chat")
	
	if _, err := bm.page.Goto("https://claude.ai/chat", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		return fmt.Errorf("failed to navigate to Claude: %v", err)
	}
	
	// Check if we're logged in
	loggedIn, err := bm.isLoggedIn()
	if err != nil {
		return err
	}
	
	if !loggedIn {
		return errors.New("not logged in to Claude, please provide valid cookies")
	}
	
	return nil
}

// isLoggedIn checks if we're logged in to Claude
func (bm *BrowserManager) isLoggedIn() (bool, error) {
	// Wait for page to load
	err := bm.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	if err != nil {
		return false, fmt.Errorf("failed waiting for page load: %v", err)
	}
	
	// Check for login elements vs chat elements
	hasLoginButton, err := bm.page.Locator("text=Log in").Count()
	if err != nil {
		return false, fmt.Errorf("failed to check for login button: %v", err)
	}
	
	if hasLoginButton > 0 {
		return false, nil
	}
	
	// Check for chat input which indicates we're logged in
	hasChatInput, err := bm.page.Locator("textarea").Count()
	if err != nil {
		return false, fmt.Errorf("failed to check for chat input: %v", err)
	}
	
	return hasChatInput > 0, nil
}

// SendMessage sends a message to Claude and returns the response
func (bm *BrowserManager) SendMessage(ctx context.Context, message string) (string, error) {
	bm.logger.Println("Sending message to Claude")
	
	// Find and click the textarea
	textarea, err := bm.page.Locator("textarea").First()
	if err != nil {
		return "", fmt.Errorf("failed to find textarea: %v", err)
	}
	
	if err := textarea.Click(); err != nil {
		return "", fmt.Errorf("failed to click textarea: %v", err)
	}
	
	// Clear existing text and type new message
	if err := textarea.Fill(""); err != nil {
		return "", fmt.Errorf("failed to clear textarea: %v", err)
	}
	
	if err := textarea.Type(message, playwright.LocatorTypeOptions{
		Delay: playwright.Float(10),
	}); err != nil {
		return "", fmt.Errorf("failed to type message: %v", err)
	}
	
	// Press Enter to send
	if err := textarea.Press("Enter"); err != nil {
		return "", fmt.Errorf("failed to press Enter: %v", err)
	}
	
	// Wait for response to appear
	// Claude's response usually appears in a div with role="article"
	bm.logger.Println("Waiting for Claude to respond")
	
	// Wait for the typing indicator to appear first
	typingIndicator, err := bm.page.Locator(".typing-indicator, .animate-pulse").First()
	if err == nil {
		// Wait for it to be visible
		err = typingIndicator.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		})
		if err == nil {
			bm.logger.Println("Typing indicator visible, Claude is generating response")
		}
	}
	
	// Now wait for the typing indicator to disappear
	err = bm.waitForClaudeToFinishTyping(ctx)
	if err != nil {
		return "", fmt.Errorf("error waiting for Claude to finish: %v", err)
	}
	
	// Extract Claude's response
	response, err := bm.extractLatestResponse()
	if err != nil {
		return "", err
	}
	
	bm.logger.Println("Successfully received response from Claude")
	return response, nil
}

// waitForClaudeToFinishTyping waits for Claude to finish generating a response
func (bm *BrowserManager) waitForClaudeToFinishTyping(ctx context.Context) error {
	timeout := 60 * time.Second
	checkInterval := 1 * time.Second
	
	startTime := time.Now()
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(startTime) > timeout {
				return errors.New("timeout waiting for Claude to finish responding")
			}
			
			// Check if typing indicators are present
			count, err := bm.page.Locator(".typing-indicator, .animate-pulse").Count()
			if err != nil {
				return fmt.Errorf("error checking typing indicators: %v", err)
			}
			
			if count == 0 {
				// Double-check after a short delay
				time.Sleep(2 * time.Second)
				
				count, err = bm.page.Locator(".typing-indicator, .animate-pulse").Count()
				if err != nil {
					return fmt.Errorf("error re-checking typing indicators: %v", err)
				}
				
				if count == 0 {
					// Claude has finished responding
					return nil
				}
			}
			
			time.Sleep(checkInterval)
		}
	}
}

// extractLatestResponse extracts the latest response from Claude
func (bm *BrowserManager) extractLatestResponse() (string, error) {
	// Get all message containers
	messages, err := bm.page.Locator("div[role='article']").All()
	if err != nil {
		return "", fmt.Errorf("failed to locate message containers: %v", err)
	}
	
	if len(messages) == 0 {
		return "", errors.New("no messages found")
	}
	
	// Get the latest message (Claude's response)
	lastMessage := messages[len(messages)-1]
	
	// Extract text content
	text, err := lastMessage.TextContent()
	if err != nil {
		return "", fmt.Errorf("failed to extract message text: %v", err)
	}
	
	return text, nil
}

// TakeScreenshot takes a screenshot for debugging
func (bm *BrowserManager) TakeScreenshot(filename string) error {
	if !bm.debug {
		return nil
	}
	
	screenshotDir := "screenshots"
	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		return fmt.Errorf("failed to create screenshots directory: %v", err)
	}
	
	path := filepath.Join(screenshotDir, filename)
	
	if err := bm.page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(path),
		FullPage: playwright.Bool(true),
	}); err != nil {
		return fmt.Errorf("failed to take screenshot: %v", err)
	}
	
	return nil
}

// StreamResponse streams Claude's response as it's being generated
func (bm *BrowserManager) StreamResponse(ctx context.Context, message string, callback func(string, bool)) error {
	bm.logger.Println("Sending message to Claude with streaming")
	
	// Find and click the textarea
	textarea, err := bm.page.Locator("textarea").First()
	if err != nil {
		return fmt.Errorf("failed to find textarea: %v", err)
	}
	
	if err := textarea.Click(); err != nil {
		return fmt.Errorf("failed to click textarea: %v", err)
	}
	
	// Clear existing text and type new message
	if err := textarea.Fill(""); err != nil {
		return fmt.Errorf("failed to clear textarea: %v", err)
	}
	
	if err := textarea.Type(message, playwright.LocatorTypeOptions{
		Delay: playwright.Float(10),
	}); err != nil {
		return fmt.Errorf("failed to type message: %v", err)
	}
	
	// Press Enter to send
	if err := textarea.Press("Enter"); err != nil {
		return fmt.Errorf("failed to press Enter: %v", err)
	}
	
	// Wait for response to start appearing
	bm.logger.Println("Waiting for Claude to start responding")
	
	// Wait for the typing indicator to appear first
	typingIndicator, err := bm.page.Locator(".typing-indicator, .animate-pulse").First()
	if err == nil {
		// Wait for it to be visible
		err = typingIndicator.WaitFor(playwright.LocatorWaitForOptions{
			State: playwright.WaitForSelectorStateVisible,
			Timeout: playwright.Float(5000),
		})
		if err == nil {
			bm.logger.Println("Typing indicator visible, Claude is generating response")
		}
	}
	
	// Start monitoring for response updates
	return bm.monitorResponseStream(ctx, callback)
}

// monitorResponseStream monitors Claude's response as it's being generated and streams it
func (bm *BrowserManager) monitorResponseStream(ctx context.Context, callback func(string, bool)) error {
	timeout := 120 * time.Second
	checkInterval := 500 * time.Millisecond
	
	startTime := time.Now()
	var lastResponse string
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(startTime) > timeout {
				callback("", true) // Signal completion due to timeout
				return errors.New("timeout waiting for Claude to finish responding")
			}
			
			// Extract current response
			currentResponse, err := bm.extractLatestResponse()
			if err != nil {
				// If we can't extract the response yet, continue waiting
				time.Sleep(checkInterval)
				continue
			}
			
			// If response has changed, send the update
			if currentResponse != lastResponse {
				callback(currentResponse, false) // Not done yet
				lastResponse = currentResponse
			}
			
			// Check if typing indicators are present
			count, err := bm.page.Locator(".typing-indicator, .animate-pulse").Count()
			if err != nil {
				return fmt.Errorf("error checking typing indicators: %v", err)
			}
			
			if count == 0 {
				// Double-check after a short delay
				time.Sleep(1 * time.Second)
				
				count, err = bm.page.Locator(".typing-indicator, .animate-pulse").Count()
				if err != nil {
					return fmt.Errorf("error re-checking typing indicators: %v", err)
				}
				
				if count == 0 {
					// Claude has finished responding
					// Send one final update with the complete response
					currentResponse, err := bm.extractLatestResponse()
					if err == nil && currentResponse != lastResponse {
						callback(currentResponse, true) // Final response
					} else {
						callback("", true) // Signal completion
					}
					return nil
				}
			}
			
			time.Sleep(checkInterval)
		}
	}
}

