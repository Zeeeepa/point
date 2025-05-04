package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	
	"github.com/Zeeeepa/point/ai-web-integration-agent/api"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.json", "Path to configuration file")
	port := flag.Int("port", 8080, "Port to listen on")
	host := flag.String("host", "0.0.0.0", "Host to listen on")
	headless := flag.Bool("headless", false, "Run in headless mode")
	debug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	// Create default config file if it doesn't exist
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		log.Printf("Config file %s not found, creating default config", *configPath)
		
		// Create default config
		config := api.Config{
			Port:               *port,
			Host:               *host,
			ClaudeURL:          "https://claude.ai/chat",
			GithubCopilotURL:   "https://github.com/features/copilot",
			BrowserUserDataDir: "~/.browser-agent",
			ScreenshotDir:      "./screenshots",
			LogFile:            "./api-server.log",
			Headless:           *headless,
			DebugMode:          *debug,
		}
		
		// Create directory if it doesn't exist
		dir := filepath.Dir(*configPath)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Fatalf("Failed to create directory for config file: %v", err)
			}
		}
		
		// Write config to file
		file, err := os.Create(*configPath)
		if err != nil {
			log.Fatalf("Failed to create config file: %v", err)
		}
		defer file.Close()
		
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(config); err != nil {
			log.Fatalf("Failed to write config file: %v", err)
		}
	}

	// Create and start the server
	server, err := api.NewServer(*configPath)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("Starting AI Web Integration API Server on %s:%d\n", server.config.Host, server.config.Port)
	fmt.Println("Available models:")
	fmt.Println("- web_claude")
	fmt.Println("- web_claude/chat")
	fmt.Println("- web_copilot")
	fmt.Println("- web_copilot/github")
	fmt.Println("\nPress Ctrl+C to stop the server")

	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
