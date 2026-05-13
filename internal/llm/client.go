package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type LLMClient interface {
	Complete(ctx context.Context, system, user string) (string, error)
}

type chatClient struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewOpenAIClient(apiKey string) LLMClient {
	return &chatClient{baseURL: "https://api.openai.com/v1/chat/completions", apiKey: apiKey, model: "gpt-4o", client: &http.Client{Timeout: 30 * time.Second}}
}

func NewDeepSeekClient(apiKey string) LLMClient {
	return &chatClient{baseURL: "https://api.deepseek.com/v1/chat/completions", apiKey: apiKey, model: "deepseek-chat", client: &http.Client{Timeout: 30 * time.Second}}
}

func NewProvider(provider, openAIKey, deepSeekKey string) (LLMClient, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai":
		if openAIKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is required")
		}
		return NewOpenAIClient(openAIKey), nil
	case "deepseek":
		if deepSeekKey == "" {
			return nil, fmt.Errorf("DEEPSEEK_API_KEY is required")
		}
		return NewDeepSeekClient(deepSeekKey), nil
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}
}

func (c *chatClient) Complete(ctx context.Context, system, user string) (string, error) {
	payload, err := json.Marshal(chatRequest{Model: c.model, Messages: []chatMessage{{Role: "system", Content: system}, {Role: "user", Content: user}}})
	if err != nil {
		return "", fmt.Errorf("marshal chat request: %w", err)
	}

	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(payload))
		if err != nil {
			return "", fmt.Errorf("create llm request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		resp, err := c.client.Do(req)
		if err != nil {
			return "", fmt.Errorf("execute llm request: %w", err)
		}
		body, readErr := io.ReadAll(resp.Body)
		closeErr := resp.Body.Close()
		if readErr != nil {
			return "", fmt.Errorf("read llm response: %w", readErr)
		}
		if closeErr != nil {
			return "", fmt.Errorf("close llm response: %w", closeErr)
		}

		if resp.StatusCode >= 500 && attempt < 2 {
			if err := waitRetry(ctx, time.Duration(attempt+1)*time.Second); err != nil {
				return "", fmt.Errorf("wait llm retry: %w", err)
			}
			continue
		}
		if resp.StatusCode >= 300 {
			return "", fmt.Errorf("llm request failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		var parsed chatResponse
		if err := json.Unmarshal(body, &parsed); err != nil {
			return "", fmt.Errorf("unmarshal llm response: %w", err)
		}
		if len(parsed.Choices) == 0 {
			return "", fmt.Errorf("llm response has no choices")
		}
		return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
	}
	return "", fmt.Errorf("llm request exhausted retries")
}

func waitRetry(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return fmt.Errorf("context done: %w", ctx.Err())
	case <-t.C:
		return nil
	}
}
