package web_copilot

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

// BrowserManager handles browser automation for GitHub Copilot web interface
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
	logFile, err := os.OpenFile("copilot_browser.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	logger := log.New(logFile, "COPILOT_BROWSER: ", log.LstdFlags|log.Lshortfile)
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
			Domain: playwright.String(".github.com"),
			Path:   playwright.String("/"),
		})
	}
	
	// Set cookies in browser context
	if err := bm.context.AddCookies(playwrightCookies); err != nil {
		return fmt.Errorf("failed to set cookies: %v", err)
	}
	
	return nil
}

// NavigateToCopilotEditor navigates to GitHub Copilot editor
func (bm *BrowserManager) NavigateToCopilotEditor() error {
	bm.logger.Println("Navigating to GitHub Copilot editor")
	
	// Navigate to GitHub Copilot playground
	if _, err := bm.page.Goto("https://github.com/features/copilot", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		return fmt.Errorf("failed to navigate to GitHub Copilot: %v", err)
	}
	
	// Check if we're logged in
	loggedIn, err := bm.isLoggedIn()
	if err != nil {
		return err
	}
	
	if !loggedIn {
		return errors.New("not logged in to GitHub, please provide valid cookies")
	}
	
	// Find and click on the "Try Copilot" button
	tryButton, err := bm.page.Locator("text=Try Copilot").First()
	if err != nil {
		return fmt.Errorf("failed to find 'Try Copilot' button: %v", err)
	}
	
	if err := tryButton.Click(); err != nil {
		return fmt.Errorf("failed to click 'Try Copilot' button: %v", err)
	}
	
	// Wait for the editor to load
	err = bm.page.WaitForSelector(".monaco-editor", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		return fmt.Errorf("failed to load Copilot editor: %v", err)
	}
	
	return nil
}

// isLoggedIn checks if we're logged in to GitHub
func (bm *BrowserManager) isLoggedIn() (bool, error) {
	// Wait for page to load
	err := bm.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	if err != nil {
		return false, fmt.Errorf("failed waiting for page load: %v", err)
	}
	
	// Check for avatar which indicates we're logged in
	avatarCount, err := bm.page.Locator(".avatar, .Header-item.position-relative.mr-0 .avatar").Count()
	if err != nil {
		return false, fmt.Errorf("failed to check for avatar: %v", err)
	}
	
	return avatarCount > 0, nil
}

// SendCodeContext sends code context to Copilot and returns the suggestion
func (bm *BrowserManager) SendCodeContext(ctx context.Context, codeContext string, language string) (string, error) {
	bm.logger.Println("Sending code context to Copilot")
	
	// Set the language in the editor if provided
	if language != "" {
		err := bm.setEditorLanguage(language)
		if err != nil {
			bm.logger.Printf("Warning: Failed to set language to %s: %v", language, err)
			// Continue anyway, as this is not critical
		}
	}
	
	// Find the editor
	editor, err := bm.page.Locator(".monaco-editor").First()
	if err != nil {
		return "", fmt.Errorf("failed to find editor: %v", err)
	}
	
	// Click on the editor to focus it
	if err := editor.Click(); err != nil {
		return "", fmt.Errorf("failed to click editor: %v", err)
	}
	
	// Clear existing code
	if err := bm.page.Keyboard.Press("Control+A"); err != nil {
		return "", fmt.Errorf("failed to select all text: %v", err)
	}
	
	if err := bm.page.Keyboard.Press("Delete"); err != nil {
		return "", fmt.Errorf("failed to delete selected text: %v", err)
	}
	
	// Type the code context
	if err := bm.page.Keyboard.Type(codeContext); err != nil {
		return "", fmt.Errorf("failed to type code context: %v", err)
	}
	
	// Trigger Copilot suggestions
	if err := bm.page.Keyboard.Press("Control+Enter"); err != nil {
		return "", fmt.Errorf("failed to trigger Copilot suggestions: %v", err)
	}
	
	// Wait for suggestions to appear
	bm.logger.Println("Waiting for Copilot suggestions")
	
	// Wait for the suggestion to appear
	err = bm.page.WaitForSelector(".copilot-suggestion, .suggest-widget", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get Copilot suggestions: %v", err)
	}
	
	// Extract the suggestion
	suggestion, err := bm.extractSuggestion()
	if err != nil {
		return "", err
	}
	
	bm.logger.Println("Successfully received suggestion from Copilot")
	return suggestion, nil
}

// setEditorLanguage sets the language mode in the Monaco editor
func (bm *BrowserManager) setEditorLanguage(language string) error {
	// Map common language names to Monaco editor language IDs
	languageMap := map[string]string{
		"python":     "python",
		"javascript": "javascript",
		"typescript": "typescript",
		"java":       "java",
		"go":         "go",
		"c#":         "csharp",
		"c++":        "cpp",
		"c":          "c",
		"ruby":       "ruby",
		"php":        "php",
		"html":       "html",
		"css":        "css",
		"rust":       "rust",
		"swift":      "swift",
		"kotlin":     "kotlin",
		"scala":      "scala",
		"shell":      "shell",
		"bash":       "shell",
		"powershell": "powershell",
		"sql":        "sql",
	}
	
	// Normalize language name
	language = strings.ToLower(language)
	
	// Get Monaco language ID
	monacoLang, ok := languageMap[language]
	if !ok {
		// If not found, use the input language as is
		monacoLang = language
	}
	
	// Execute JavaScript to change the language
	_, err := bm.page.Evaluate(fmt.Sprintf(`
		try {
			const editor = monaco.editor.getEditors()[0];
			if (editor) {
				monaco.editor.setModelLanguage(editor.getModel(), "%s");
				return true;
			}
			return false;
		} catch (e) {
			console.error("Failed to set language:", e);
			return false;
		}
	`, monacoLang))
	
	return err
}

// extractSuggestion extracts the suggestion from Copilot
func (bm *BrowserManager) extractSuggestion() (string, error) {
	// Try to get the suggestion from the ghost text
	ghostText, err := bm.page.Evaluate(`
		try {
			const ghostText = document.querySelector('.suggest-widget .monaco-list-row');
			if (ghostText) {
				return ghostText.textContent;
			}
			return "";
		} catch (e) {
			console.error("Failed to extract ghost text:", e);
			return "";
		}
	`)
	
	if err != nil {
		return "", fmt.Errorf("failed to extract ghost text: %v", err)
	}
	
	if ghostText != nil && ghostText.(string) != "" {
		return ghostText.(string), nil
	}
	
	// If ghost text is not available, try to get the inline suggestion
	inlineSuggestion, err := bm.page.Evaluate(`
		try {
			const editor = monaco.editor.getEditors()[0];
			if (editor) {
				// Get the current value
				const currentValue = editor.getValue();
				
				// Accept the suggestion
				editor.trigger('keyboard', 'acceptSelectedSuggestion', null);
				
				// Get the new value after accepting the suggestion
				const newValue = editor.getValue();
				
				// Return the difference
				return newValue.substring(currentValue.length);
			}
			return "";
		} catch (e) {
			console.error("Failed to extract inline suggestion:", e);
			return "";
		}
	`)
	
	if err != nil {
		return "", fmt.Errorf("failed to extract inline suggestion: %v", err)
	}
	
	if inlineSuggestion != nil && inlineSuggestion.(string) != "" {
		return inlineSuggestion.(string), nil
	}
	
	return "No suggestion available", nil
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

// StreamSuggestion streams Copilot's suggestion as it's being generated
func (bm *BrowserManager) StreamSuggestion(ctx context.Context, codeContext string, language string, callback func(string, bool)) error {
	bm.logger.Println("Sending code context to Copilot with streaming")
	
	// Set the language in the editor if provided
	if language != "" {
		err := bm.setEditorLanguage(language)
		if err != nil {
			bm.logger.Printf("Warning: Failed to set language to %s: %v", language, err)
			// Continue anyway, as this is not critical
		}
	}
	
	// Find the editor
	editor, err := bm.page.Locator(".monaco-editor").First()
	if err != nil {
		return fmt.Errorf("failed to find editor: %v", err)
	}
	
	// Click on the editor to focus it
	if err := editor.Click(); err != nil {
		return fmt.Errorf("failed to click editor: %v", err)
	}
	
	// Clear existing code
	if err := bm.page.Keyboard.Press("Control+A"); err != nil {
		return fmt.Errorf("failed to select all text: %v", err)
	}
	
	if err := bm.page.Keyboard.Press("Delete"); err != nil {
		return fmt.Errorf("failed to delete selected text: %v", err)
	}
	
	// Type the code context
	if err := bm.page.Keyboard.Type(codeContext); err != nil {
		return fmt.Errorf("failed to type code context: %v", err)
	}
	
	// Trigger Copilot suggestions
	if err := bm.page.Keyboard.Press("Control+Enter"); err != nil {
		return fmt.Errorf("failed to trigger Copilot suggestions: %v", err)
	}
	
	// Wait for suggestions to appear
	bm.logger.Println("Waiting for Copilot suggestions")
	
	// Wait for the suggestion to appear
	err = bm.page.WaitForSelector(".copilot-suggestion, .suggest-widget", playwright.PageWaitForSelectorOptions{
		State: playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		return fmt.Errorf("failed to get Copilot suggestions: %v", err)
	}
	
	// Start monitoring for suggestion updates
	return bm.monitorSuggestionStream(ctx, callback)
}

// monitorSuggestionStream monitors Copilot's suggestion as it's being generated and streams it
func (bm *BrowserManager) monitorSuggestionStream(ctx context.Context, callback func(string, bool)) error {
	timeout := 30 * time.Second
	checkInterval := 500 * time.Millisecond
	
	startTime := time.Now()
	var lastSuggestion string
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if time.Since(startTime) > timeout {
				callback("", true) // Signal completion due to timeout
				return errors.New("timeout waiting for Copilot to finish suggesting")
			}
			
			// Extract current suggestion
			currentSuggestion, err := bm.extractSuggestion()
			if err != nil {
				// If we can't extract the suggestion yet, continue waiting
				time.Sleep(checkInterval)
				continue
			}
			
			// If suggestion has changed, send the update
			if currentSuggestion != lastSuggestion {
				callback(currentSuggestion, false) // Not done yet
				lastSuggestion = currentSuggestion
			}
			
			// Check if suggestion is still being generated
			isGenerating, err := bm.page.Evaluate(`
				try {
					const ghostText = document.querySelector('.suggest-widget');
					return ghostText && ghostText.style.display !== 'none';
				} catch (e) {
					return false;
				}
			`)
			
			if err != nil || isGenerating == nil || !isGenerating.(bool) {
				// Double-check after a short delay
				time.Sleep(1 * time.Second)
				
				isGenerating, err = bm.page.Evaluate(`
					try {
						const ghostText = document.querySelector('.suggest-widget');
						return ghostText && ghostText.style.display !== 'none';
					} catch (e) {
						return false;
					}
				`)
				
				if err != nil || isGenerating == nil || !isGenerating.(bool) {
					// Copilot has finished suggesting
					// Send one final update with the complete suggestion
					currentSuggestion, err := bm.extractSuggestion()
					if err == nil && currentSuggestion != lastSuggestion {
						callback(currentSuggestion, true) // Final suggestion
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

