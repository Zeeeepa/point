package api

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

// NavigateToCopilotEditor navigates to the GitHub Copilot editor
func (s *Session) NavigateToCopilotEditor() error {
	s.logger.Println("Navigating to GitHub Copilot editor")
	
	// Navigate to GitHub Copilot
	if err := chromedp.Run(s.ctx, chromedp.Navigate(s.config.GithubCopilotURL)); err != nil {
		return fmt.Errorf("failed to navigate to GitHub Copilot: %v", err)
	}

	// Wait for page to load
	if err := chromedp.Run(s.ctx, 
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
	); err != nil {
		return fmt.Errorf("failed waiting for GitHub Copilot page: %v", err)
	}

	// Find and click on the "Try Copilot" button
	var tryButtonExists bool
	err := chromedp.Run(s.ctx, chromedp.Evaluate(`
		document.querySelector('a[href*="copilot/editor"]') !== null ||
		document.querySelector('a:contains("Try Copilot")') !== null
	`, &tryButtonExists))
	
	if err != nil {
		return fmt.Errorf("failed to check for 'Try Copilot' button: %v", err)
	}

	if tryButtonExists {
		// Click the button
		err := chromedp.Run(s.ctx, chromedp.Click(`a[href*="copilot/editor"], a:contains("Try Copilot")`, chromedp.ByQuery))
		if err != nil {
			return fmt.Errorf("failed to click 'Try Copilot' button: %v", err)
		}
	}

	// Wait for the editor to load
	err = chromedp.Run(s.ctx, 
		chromedp.WaitVisible(`.monaco-editor`, chromedp.ByQuery),
	)
	if err != nil {
		return fmt.Errorf("failed waiting for Copilot editor: %v", err)
	}

	return nil
}

// SendCodeContextToCopilot sends code context to Copilot
func (s *Session) SendCodeContextToCopilot(codeContext string, language string) error {
	s.logger.Println("Sending code context to Copilot")
	
	// Set the language in the editor if provided
	if language != "" {
		err := chromedp.Run(s.ctx, chromedp.Evaluate(fmt.Sprintf(`
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
		`, language), nil))
		
		if err != nil {
			s.logger.Printf("Warning: Failed to set language to %s: %v", language, err)
			// Continue anyway, as this is not critical
		}
	}
	
	// Find the editor
	err := chromedp.Run(s.ctx,
		chromedp.Click(`.monaco-editor`, chromedp.ByQuery),
		chromedp.KeyEvent("Control+a"), // Select all
		chromedp.KeyEvent("Delete"), // Delete selected
		chromedp.SendKeys(`.monaco-editor`, codeContext, chromedp.ByQuery),
	)
	
	if err != nil {
		return fmt.Errorf("failed to input code context: %v", err)
	}
	
	// Trigger Copilot suggestions
	err = chromedp.Run(s.ctx,
		chromedp.KeyEvent("Control+Enter"), // This may vary based on the actual trigger
	)
	
	if err != nil {
		return fmt.Errorf("failed to trigger Copilot suggestions: %v", err)
	}
	
	// Wait for suggestions to appear
	time.Sleep(3 * time.Second)
	
	return nil
}

// ExtractCopilotSuggestion extracts the suggestion from Copilot
func (s *Session) ExtractCopilotSuggestion() (string, error) {
	var suggestion string
	
	// Try to get the suggestion from the ghost text
	err := chromedp.Run(s.ctx, chromedp.Evaluate(`
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
	`, &suggestion))
	
	if err != nil {
		return "", fmt.Errorf("failed to extract ghost text: %v", err)
	}
	
	if suggestion != "" {
		return suggestion, nil
	}
	
	// If ghost text is not available, try to get the inline suggestion
	err = chromedp.Run(s.ctx, chromedp.Evaluate(`
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
	`, &suggestion))
	
	if err != nil {
		return "", fmt.Errorf("failed to extract inline suggestion: %v", err)
	}
	
	if suggestion != "" {
		return suggestion, nil
	}
	
	return "No suggestion available", nil
}

// IsCopilotGenerating checks if Copilot is still generating suggestions
func (s *Session) IsCopilotGenerating() (bool, error) {
	var isGenerating bool
	
	err := chromedp.Run(s.ctx, chromedp.Evaluate(`
		try {
			const ghostText = document.querySelector('.suggest-widget');
			return ghostText && ghostText.style.display !== 'none';
		} catch (e) {
			return false;
		}
	`, &isGenerating))
	
	if err != nil {
		return false, fmt.Errorf("failed to check if Copilot is generating: %v", err)
	}
	
	return isGenerating, nil
}

// SendPromptToClaude sends a prompt to Claude
func (s *Session) SendPromptToClaude(prompt string) error {
	s.logger.Println("Sending prompt to Claude")
	
	// Navigate to Claude
	if err := chromedp.Run(s.ctx, chromedp.Navigate(s.config.ClaudeURL)); err != nil {
		return fmt.Errorf("failed to navigate to Claude: %v", err)
	}

	// Wait for Claude to load
	if err := chromedp.Run(s.ctx, 
		chromedp.WaitVisible(`textarea`, chromedp.ByQuery),
	); err != nil {
		return fmt.Errorf("failed waiting for Claude input: %v", err)
	}

	// Clear existing text and type new prompt
	if err := chromedp.Run(s.ctx,
		chromedp.Click(`textarea`, chromedp.ByQuery),
		chromedp.KeyEvent("Control+a"), // Select all
		chromedp.KeyEvent("Delete"), // Delete selected
		chromedp.SendKeys(`textarea`, prompt, chromedp.ByQuery),
		chromedp.KeyEvent("Enter"), // Send the prompt
	); err != nil {
		return fmt.Errorf("failed to input prompt: %v", err)
	}

	// Wait for response to start appearing
	time.Sleep(2 * time.Second)
	
	return nil
}

// ExtractClaudeResponse extracts the response from Claude
func (s *Session) ExtractClaudeResponse() (string, error) {
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

// IsClaudeTyping checks if Claude is still typing
func (s *Session) IsClaudeTyping() (bool, error) {
	var isTyping bool
	
	err := chromedp.Run(s.ctx, chromedp.Evaluate(`
		document.querySelector('.typing-indicator') !== null || 
		document.querySelector('.animate-pulse') !== null
	`, &isTyping))
	
	if err != nil {
		return false, fmt.Errorf("failed to check if Claude is typing: %v", err)
	}
	
	return isTyping, nil
}

// WaitForClaudeToFinishTyping waits for Claude to finish typing
func (s *Session) WaitForClaudeToFinishTyping(ctx context.Context) error {
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
			isTyping, err := s.IsClaudeTyping()
			if err != nil {
				return fmt.Errorf("error checking typing indicators: %v", err)
			}
			
			if !isTyping {
				// Double-check after a short delay
				time.Sleep(2 * time.Second)
				
				isTyping, err = s.IsClaudeTyping()
				if err != nil {
					return fmt.Errorf("error re-checking typing indicators: %v", err)
				}
				
				if !isTyping {
					// Claude has finished responding
					return nil
				}
			}
			
			time.Sleep(checkInterval)
		}
	}
}

