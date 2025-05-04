package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Configuration for the API server
type Config struct {
	Port               int    `json:"port"`
	Host               string `json:"host"`
	ClaudeURL          string `json:"claude_url"`
	GithubCopilotURL   string `json:"github_copilot_url"`
	BrowserUserDataDir string `json:"browser_user_data_dir"`
	ScreenshotDir      string `json:"screenshot_dir"`
	LogFile            string `json:"log_file"`
	Headless           bool   `json:"headless"`
	DebugMode          bool   `json:"debug_mode"`
}

// Server represents the API server
type Server struct {
	config      Config
	logger      *log.Logger
	claudeAgent *ClaudeAgent
	copilotAgent *CopilotAgent
	mu          sync.Mutex
}

// NewServer creates a new API server
func NewServer(configPath string) (*Server, error) {
	// Load configuration
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// Setup logging
	logFile, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	logger := log.New(logFile, "API_SERVER: ", log.LstdFlags|log.Lshortfile)
	logger.Println("Initializing API server")

	// Create screenshots directory if it doesn't exist
	if err := os.MkdirAll(config.ScreenshotDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create screenshots directory: %v", err)
	}

	// Initialize Claude agent
	claudeAgent, err := NewClaudeAgent(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Claude agent: %v", err)
	}

	// Initialize Copilot agent
	copilotAgent, err := NewCopilotAgent(config)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Copilot agent: %v", err)
	}

	return &Server{
		config:      config,
		logger:      logger,
		claudeAgent: claudeAgent,
		copilotAgent: copilotAgent,
	}, nil
}

// Start starts the API server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.logger.Printf("Starting API server on %s", addr)

	// Setup routes
	http.HandleFunc("/v1/chat/completions", s.handleChatCompletions)
	http.HandleFunc("/v1/completions", s.handleCompletions)
	http.HandleFunc("/v1/models", s.handleModels)
	http.HandleFunc("/health", s.handleHealth)

	// Start server
	return http.ListenAndServe(addr, nil)
}

// handleChatCompletions handles chat completions requests
func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	s.logger.Printf("Received chat completions request: %s %s", r.Method, r.URL.Path)

	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req ChatCompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.Messages) == 0 {
		http.Error(w, "Messages cannot be empty", http.StatusBadRequest)
		return
	}

	// Process request based on model
	var resp interface{}
	var err error

	if strings.HasPrefix(req.Model, "web_claude") {
		// Handle Claude request
		resp, err = s.claudeAgent.ProcessChatCompletion(r.Context(), req)
	} else if strings.HasPrefix(req.Model, "web_copilot") {
		// Handle Copilot request
		resp, err = s.copilotAgent.ProcessChatCompletion(r.Context(), req)
	} else {
		http.Error(w, fmt.Sprintf("Unsupported model: %s", req.Model), http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing request: %v", err), http.StatusInternalServerError)
		return
	}

	// Handle streaming response
	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Transfer-Encoding", "chunked")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		// Send streaming response
		for chunk := range resp.(chan StreamChunk) {
			data, err := json.Marshal(chunk.Data)
			if err != nil {
				s.logger.Printf("Error marshaling chunk: %v", err)
				continue
			}

			fmt.Fprintf(w, "data: %s\n\n", string(data))
			flusher.Flush()

			if chunk.Done {
				fmt.Fprintf(w, "data: [DONE]\n\n")
				flusher.Flush()
				break
			}
		}
	} else {
		// Send non-streaming response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Printf("Error encoding response: %v", err)
		}
	}
}

// handleCompletions handles completions requests
func (s *Server) handleCompletions(w http.ResponseWriter, r *http.Request) {
	s.logger.Printf("Received completions request: %s %s", r.Method, r.URL.Path)

	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Prompt == "" {
		http.Error(w, "Prompt cannot be empty", http.StatusBadRequest)
		return
	}

	// Process request based on model
	var resp interface{}
	var err error

	if strings.HasPrefix(req.Model, "web_claude") {
		// Handle Claude request
		resp, err = s.claudeAgent.ProcessCompletion(r.Context(), req)
	} else if strings.HasPrefix(req.Model, "web_copilot") {
		// Handle Copilot request
		resp, err = s.copilotAgent.ProcessCompletion(r.Context(), req)
	} else {
		http.Error(w, fmt.Sprintf("Unsupported model: %s", req.Model), http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing request: %v", err), http.StatusInternalServerError)
		return
	}

	// Handle streaming response
	if req.Stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Transfer-Encoding", "chunked")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		// Send streaming response
		for chunk := range resp.(chan StreamChunk) {
			data, err := json.Marshal(chunk.Data)
			if err != nil {
				s.logger.Printf("Error marshaling chunk: %v", err)
				continue
			}

			fmt.Fprintf(w, "data: %s\n\n", string(data))
			flusher.Flush()

			if chunk.Done {
				fmt.Fprintf(w, "data: [DONE]\n\n")
				flusher.Flush()
				break
			}
		}
	} else {
		// Send non-streaming response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			s.logger.Printf("Error encoding response: %v", err)
		}
	}
}

// handleModels handles models requests
func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	s.logger.Printf("Received models request: %s %s", r.Method, r.URL.Path)

	// Only accept GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Return available models
	models := ModelsResponse{
		Object: "list",
		Data: []Model{
			{
				ID:      "web_claude",
				Object:  "model",
				Created: time.Now().Unix(),
				OwnedBy: "anthropic",
			},
			{
				ID:      "web_claude/chat",
				Object:  "model",
				Created: time.Now().Unix(),
				OwnedBy: "anthropic",
			},
			{
				ID:      "web_copilot",
				Object:  "model",
				Created: time.Now().Unix(),
				OwnedBy: "github",
			},
			{
				ID:      "web_copilot/github",
				Object:  "model",
				Created: time.Now().Unix(),
				OwnedBy: "github",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(models); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// loadConfig loads the configuration from a file
func loadConfig(path string) (Config, error) {
	// Default configuration
	config := Config{
		Port:               8080,
		Host:               "0.0.0.0",
		ClaudeURL:          "https://claude.ai/chat",
		GithubCopilotURL:   "https://github.com/features/copilot",
		BrowserUserDataDir: "~/.browser-agent",
		ScreenshotDir:      "./screenshots",
		LogFile:            "./api-server.log",
		Headless:           true,
		DebugMode:          false,
	}

	// If no config file specified, return defaults
	if path == "" {
		return config, nil
	}

	// Read the configuration file
	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse the JSON
	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}

// StreamChunk represents a chunk of a streaming response
type StreamChunk struct {
	Data interface{}
	Done bool
}

// Request and response types

// ChatCompletionRequest represents a chat completion request
type ChatCompletionRequest struct {
	Model       string                  `json:"model"`
	Messages    []ChatCompletionMessage `json:"messages"`
	Temperature float64                 `json:"temperature,omitempty"`
	MaxTokens   int                     `json:"max_tokens,omitempty"`
	Stream      bool                    `json:"stream,omitempty"`
}

// ChatCompletionMessage represents a message in a chat completion request
type ChatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents a chat completion response
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ChatCompletionStreamResponse represents a streaming chat completion response
type ChatCompletionStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int               `json:"index"`
		Delta        map[string]string `json:"delta"`
		FinishReason *string           `json:"finish_reason"`
	} `json:"choices"`
}

// CompletionRequest represents a completion request
type CompletionRequest struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	Temperature float64  `json:"temperature,omitempty"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Stop        []string `json:"stop,omitempty"`
	Stream      bool     `json:"stream,omitempty"`
}

// CompletionResponse represents a completion response
type CompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string `json:"text"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// CompletionStreamResponse represents a streaming completion response
type CompletionStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string  `json:"text"`
		Index        int     `json:"index"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// ModelsResponse represents a models response
type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// Model represents a model in a models response
type Model struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

