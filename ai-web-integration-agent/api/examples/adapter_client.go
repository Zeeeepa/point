package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

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

// CompletionRequest represents a completion request
type CompletionRequest struct {
	Model       string   `json:"model"`
	Prompt      string   `json:"prompt"`
	Temperature float64  `json:"temperature,omitempty"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Stop        []string `json:"stop,omitempty"`
	Stream      bool     `json:"stream,omitempty"`
}

func main() {
	// Parse command line flags
	url := flag.String("url", "http://localhost:8080", "API server URL")
	model := flag.String("model", "web_claude", "Model to use (web_claude, web_claude/chat, web_copilot, web_copilot/github)")
	prompt := flag.String("prompt", "Hello, how are you?", "Prompt to send")
	stream := flag.Bool("stream", false, "Use streaming API")
	flag.Parse()

	// Create HTTP client
	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	// Determine endpoint and request body based on model
	var endpoint string
	var reqBody []byte
	var err error

	if strings.HasPrefix(*model, "web_claude") {
		endpoint = fmt.Sprintf("%s/v1/chat/completions", *url)
		req := ChatCompletionRequest{
			Model: *model,
			Messages: []ChatCompletionMessage{
				{
					Role:    "user",
					Content: *prompt,
				},
			},
			Stream: *stream,
		}
		reqBody, err = json.Marshal(req)
	} else {
		endpoint = fmt.Sprintf("%s/v1/completions", *url)
		req := CompletionRequest{
			Model:  *model,
			Prompt: *prompt,
			Stream: *stream,
		}
		reqBody, err = json.Marshal(req)
	}

	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		os.Exit(1)
	}

	// Create request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	fmt.Printf("Sending request to %s...\n", endpoint)
	fmt.Printf("Request data: %s\n", string(reqBody))

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Error: %d %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	// Handle response
	if *stream {
		// Handle streaming response
		fmt.Println("Response:")
		handleStreamingResponse(resp.Body)
	} else {
		// Handle non-streaming response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Response: %s\n", string(body))
	}
}

// handleStreamingResponse handles a streaming response
func handleStreamingResponse(body io.Reader) {
	reader := NewSSEReader(body)
	for {
		event, err := reader.ReadEvent()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Error reading event: %v\n", err)
			break
		}

		if event.Data == "[DONE]" {
			break
		}

		fmt.Print(event.Data)
	}
	fmt.Println()
}

// SSEReader is a reader for Server-Sent Events
type SSEReader struct {
	reader io.Reader
	buffer []byte
}

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	Event string
	Data  string
}

// NewSSEReader creates a new SSEReader
func NewSSEReader(reader io.Reader) *SSEReader {
	return &SSEReader{
		reader: reader,
		buffer: make([]byte, 0),
	}
}

// ReadEvent reads a Server-Sent Event
func (r *SSEReader) ReadEvent() (*SSEEvent, error) {
	event := &SSEEvent{}
	for {
		line, err := r.readLine()
		if err != nil {
			return nil, err
		}

		if line == "" {
			// End of event
			return event, nil
		}

		if strings.HasPrefix(line, "event:") {
			event.Event = strings.TrimSpace(line[6:])
		} else if strings.HasPrefix(line, "data:") {
			event.Data = strings.TrimSpace(line[5:])
		}
	}
}

// readLine reads a line from the reader
func (r *SSEReader) readLine() (string, error) {
	for {
		// Check if we have a line in the buffer
		i := bytes.IndexByte(r.buffer, '\n')
		if i >= 0 {
			line := string(r.buffer[:i])
			r.buffer = r.buffer[i+1:]
			return line, nil
		}

		// Read more data
		buf := make([]byte, 1024)
		n, err := r.reader.Read(buf)
		if err != nil {
			if err == io.EOF && len(r.buffer) > 0 {
				// Return the last line
				line := string(r.buffer)
				r.buffer = nil
				return line, nil
			}
			return "", err
		}

		r.buffer = append(r.buffer, buf[:n]...)
	}
}

