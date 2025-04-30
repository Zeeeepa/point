package main

import (
	"bufio"
	"encoding/json"
	"flag"
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

// Config holds the configuration for the login process
type Config struct {
	Service     string // "claude" or "github"
	URL         string
	UserDataDir string
	OutputFile  string
	Debug       bool
}

func main() {
	// Parse command line arguments
	service := flag.String("service", "", "Service to log in to (claude or github)")
	outputFile := flag.String("output", "", "Output file for cookies (default: <service>_cookies.json)")
	userDataDir := flag.String("user-data-dir", "", "User data directory (default: ~/.browser-<service>)")
	debug := flag.Bool("debug", false, "Enable debug mode")

	flag.Parse()

	// Validate service
	if *service != "claude" && *service != "github" {
		fmt.Println("Error: Service must be either 'claude' or 'github'")
		flag.Usage()
		os.Exit(1)
	}

	// Set defaults if not provided
	if *userDataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		*userDataDir = filepath.Join(homeDir, fmt.Sprintf(".browser-%s", *service))
	} else if strings.HasPrefix(*userDataDir, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		*userDataDir = filepath.Join(homeDir, (*userDataDir)[1:])
	}

	if *outputFile == "" {
		*outputFile = fmt.Sprintf("%s_cookies.json", *service)
	}

	// Create config
	config := Config{
		Service:     *service,
		UserDataDir: *userDataDir,
		OutputFile:  *outputFile,
		Debug:       *debug,
	}

	// Set URL based on service
	switch config.Service {
	case "claude":
		config.URL = "https://claude.ai/chat"
	case "github":
		config.URL = "https://github.com/login"
	}

	// Ensure Playwright is installed
	if err := ensurePlaywrightInstalled(); err != nil {
		log.Fatalf("Failed to install Playwright: %v", err)
	}

	// Run the login process
	if err := runLoginProcess(config); err != nil {
		log.Fatalf("Login process failed: %v", err)
	}

	fmt.Printf("Successfully saved cookies to %s\n", config.OutputFile)
}

// ensurePlaywrightInstalled makes sure Playwright is installed
func ensurePlaywrightInstalled() error {
	// Check if playwright is installed by running a simple command
	cmd := exec.Command("playwright", "version")
	if err := cmd.Run(); err != nil {
		// If not installed, install it
		fmt.Println("Playwright not found, installing...")
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

// runLoginProcess handles the interactive login process
func runLoginProcess(config Config) error {
	fmt.Printf("Starting interactive login for %s...\n", config.Service)
	
	// Create user data directory if it doesn't exist
	if err := os.MkdirAll(config.UserDataDir, 0755); err != nil {
		return fmt.Errorf("failed to create user data directory: %v", err)
	}

	// Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("could not start playwright: %v", err)
	}
	defer pw.Stop()

	// Launch browser (non-headless for interactive login)
	browserOpts := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
		Args: []string{
			"--disable-web-security",
			"--disable-features=IsolateOrigins,site-per-process",
		},
	}

	browser, err := pw.Chromium.Launch(browserOpts)
	if err != nil {
		return fmt.Errorf("could not launch browser: %v", err)
	}
	defer browser.Close()

	// Create browser context with user data directory
	contextOpts := playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	}

	context, err := browser.NewContext(contextOpts)
	if err != nil {
		return fmt.Errorf("could not create browser context: %v", err)
	}
	defer context.Close()

	// Create page
	page, err := context.NewPage()
	if err != nil {
		return fmt.Errorf("could not create page: %v", err)
	}

	// Set default timeout
	page.SetDefaultTimeout(30000)

	// Navigate to the service URL
	fmt.Printf("Opening %s in browser...\n", config.URL)
	if _, err := page.Goto(config.URL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		return fmt.Errorf("failed to navigate to %s: %v", config.URL, err)
	}

	// Prompt user to log in
	fmt.Println("Please log in to the service in the browser window that just opened.")
	fmt.Println("Once you are logged in, type 'Y' and press Enter to continue...")

	reader := bufio.NewReader(os.Stdin)
	for {
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if strings.ToUpper(input) == "Y" {
			break
		}
		fmt.Println("Please type 'Y' when you have logged in...")
	}

	// Extract cookies
	cookies, err := context.Cookies()
	if err != nil {
		return fmt.Errorf("failed to extract cookies: %v", err)
	}

	// Convert cookies to a map for storage
	cookieMap := make(map[string]string)
	for _, cookie := range cookies {
		cookieMap[cookie.Name] = cookie.Value
	}

	// Save cookies to file
	cookieJSON, err := json.MarshalIndent(cookieMap, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cookies to JSON: %v", err)
	}

	if err := os.WriteFile(config.OutputFile, cookieJSON, 0644); err != nil {
		return fmt.Errorf("failed to write cookies to file: %v", err)
	}

	return nil
}

