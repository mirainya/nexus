package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type AnthropicProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewAnthropic(apiKey, baseURL string) *AnthropicProvider {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	return &AnthropicProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 300 * time.Second},
	}
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

func (p *AnthropicProvider) Chat(ctx context.Context, req Request) (*Response, error) {
	var system string
	var messages []map[string]any
	for _, m := range req.Messages {
		if m.Role == "system" {
			if s, ok := m.Content.(string); ok {
				system = s
			}
			continue
		}
		switch c := m.Content.(type) {
		case []ContentPart:
			var parts []map[string]any
			for _, cp := range c {
				switch cp.Type {
				case "image_url":
					if cp.ImageURL != nil {
						parts = append(parts, map[string]any{
							"type":   "image",
							"source": map[string]any{"type": "url", "url": cp.ImageURL.URL},
						})
					}
				default:
					parts = append(parts, map[string]any{"type": "text", "text": cp.Text})
				}
			}
			messages = append(messages, map[string]any{"role": m.Role, "content": parts})
		default:
			messages = append(messages, map[string]any{"role": m.Role, "content": c})
		}
	}

	body := map[string]any{
		"model":      req.Model,
		"messages":   messages,
		"max_tokens": req.MaxTokens,
	}
	if system != "" {
		body["system"] = system
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.MaxTokens <= 0 {
		body["max_tokens"] = 4096
	}

	data, _ := json.Marshal(body)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/messages", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
		return nil, &RateLimitError{Provider: "anthropic", RetryAfter: retryAfter}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic api error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Model string `json:"model"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if len(result.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	return &Response{
		Content: result.Content[0].Text,
		Model:   result.Model,
		Usage: Usage{
			PromptTokens:     result.Usage.InputTokens,
			CompletionTokens: result.Usage.OutputTokens,
			TotalTokens:      result.Usage.InputTokens + result.Usage.OutputTokens,
		},
	}, nil
}
