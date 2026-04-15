package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cognitree/backend/internal/infrastructure/config"
)

type OpenAIClient struct {
	baseURL    string
	apiKey     string
	model      string
	maxTokens  int
	httpClient *http.Client
}

func NewOpenAIClient(cfg config.AIConfig) *OpenAIClient {
	return &OpenAIClient{
		baseURL:    cfg.BaseURL,
		apiKey:     cfg.APIKey,
		model:      cfg.Model,
		maxTokens:  cfg.MaxTokens,
		httpClient: &http.Client{},
	}
}

type chatRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens,omitempty"`
	Stream    bool          `json:"stream,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type chatStreamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *OpenAIClient) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if c.apiKey == "" {
		return c.mockChat(systemPrompt, userPrompt), nil
	}

	messages := []chatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	reqBody := chatRequest{
		Model:     c.model,
		Messages:  messages,
		MaxTokens: c.maxTokens,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("api error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) ChatStream(ctx context.Context, systemPrompt, userPrompt string, onDelta func(string) error) error {
	if c.apiKey == "" {
		return c.mockChatStream(systemPrompt, userPrompt, onDelta)
	}

	messages := []chatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	reqBody := chatRequest{
		Model:     c.model,
		Messages:  messages,
		MaxTokens: c.maxTokens,
		Stream:    true,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		respBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("read error response: %w", readErr)
		}
		return fmt.Errorf("api error: %s", strings.TrimSpace(string(respBytes)))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, readErr := reader.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			return fmt.Errorf("read stream: %w", readErr)
		}

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				return nil
			}

			var chunk chatStreamResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				return fmt.Errorf("unmarshal stream chunk: %w", err)
			}

			if chunk.Error != nil {
				return fmt.Errorf("api error: %s", chunk.Error.Message)
			}
			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta.Content
				if delta != "" {
					if err := onDelta(delta); err != nil {
						return err
					}
				}
			}
		}

		if readErr == io.EOF {
			return nil
		}
	}
}

func (c *OpenAIClient) mockChat(systemPrompt, userPrompt string) string {
	return fmt.Sprintf(
		"**[Mock Mode]** AI API Key is not configured.\n\n"+
			"**System Prompt**:\n%s\n\n"+
			"Your question looks promising. Here is a lightweight simulated response:\n\n"+
			"1. **Core idea**: understand the tree goal before answering the current ask.\n"+
			"2. **Deeper angle**: preserve anchor evidence so branch intent stays clear.\n"+
			"3. **Practical advice**: start from the current path, then expand only if needed.\n\n"+
			"> Tip: configure `ai.api_key` in `config.yaml` to get real model output.\n\n"+
			"---\n*Received question: %s*",
		systemPrompt,
		userPrompt,
	)
}

func (c *OpenAIClient) mockChatStream(systemPrompt, userPrompt string, onDelta func(string) error) error {
	answer := c.mockChat(systemPrompt, userPrompt)
	runes := []rune(answer)
	const chunkSize = 48

	for start := 0; start < len(runes); start += chunkSize {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}

		if err := onDelta(string(runes[start:end])); err != nil {
			return err
		}
		time.Sleep(8 * time.Millisecond)
	}

	return nil
}
