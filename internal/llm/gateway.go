package llm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/pkg/config"
	"github.com/mirainya/nexus/pkg/logger"
	"github.com/mirainya/nexus/pkg/metrics"
	"go.uber.org/zap"
)

type Gateway struct {
	providers  map[string]Provider
	defaults   map[string]string
	pricing    map[string][2]float64
	semaphores map[string]chan struct{}
	fallback   string
	mu         sync.RWMutex
}

var G *Gateway

func Init() {
	G = &Gateway{
		providers:  make(map[string]Provider),
		defaults:   make(map[string]string),
		pricing:    make(map[string][2]float64),
		semaphores: make(map[string]chan struct{}),
	}
	G.LoadFromDB()
}

func (g *Gateway) LoadFromDB() {
	g.mu.Lock()
	defer g.mu.Unlock()

	var list []model.LLMProvider
	model.DB().Where("active = ?", true).Find(&list)

	g.providers = make(map[string]Provider)
	g.defaults = make(map[string]string)
	g.pricing = make(map[string][2]float64)
	g.semaphores = make(map[string]chan struct{})
	g.fallback = ""

	for _, p := range list {
		if p.APIKey == "" {
			continue
		}
		switch p.Name {
		case "openai":
			g.providers[p.Name] = NewOpenAI(p.APIKey, p.BaseURL)
		case "anthropic":
			g.providers[p.Name] = NewAnthropic(p.APIKey, p.BaseURL)
		case "doubao":
			g.providers[p.Name] = NewDoubao(p.APIKey, p.BaseURL)
		default:
			g.providers[p.Name] = NewOpenAI(p.APIKey, p.BaseURL)
		}
		g.defaults[p.Name] = p.DefaultModel
		g.pricing[p.Name] = [2]float64{p.InputPrice, p.OutputPrice}

		maxConc := p.MaxConcurrency
		if maxConc <= 0 {
			maxConc = 10
		}
		g.semaphores[p.Name] = make(chan struct{}, maxConc)

		if p.IsDefault {
			g.fallback = p.Name
		}
		logger.Info("llm provider registered",
			zap.String("provider", p.Name),
			zap.Int("max_concurrency", maxConc))
	}

	if g.fallback == "" && len(g.providers) > 0 {
		for name := range g.providers {
			g.fallback = name
			break
		}
	}
}

func (g *Gateway) acquire(ctx context.Context, providerName string) error {
	g.mu.RLock()
	sem, ok := g.semaphores[providerName]
	g.mu.RUnlock()
	if !ok {
		return nil
	}
	select {
	case sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (g *Gateway) release(providerName string) {
	g.mu.RLock()
	sem, ok := g.semaphores[providerName]
	g.mu.RUnlock()
	if !ok {
		return
	}
	select {
	case <-sem:
	default:
	}
}

func retryBackoff(attempt int, err error) time.Duration {
	var rlErr *RateLimitError
	if errors.As(err, &rlErr) && rlErr.RetryAfter > 0 {
		return time.Duration(rlErr.RetryAfter) * time.Second
	}
	base := time.Duration(1<<uint(attempt)) * time.Second
	if base > 30*time.Second {
		base = 30 * time.Second
	}
	return base
}

func (g *Gateway) Chat(ctx context.Context, req Request) (*Response, error) {
	g.mu.RLock()
	providerName := req.Provider
	if providerName == "" {
		providerName = g.fallback
	}

	p, ok := g.providers[providerName]
	if !ok {
		g.mu.RUnlock()
		return nil, fmt.Errorf("provider not available: %s", providerName)
	}

	if req.Model == "" {
		req.Model = g.defaults[providerName]
	}
	g.mu.RUnlock()

	if err := g.acquire(ctx, providerName); err != nil {
		return nil, fmt.Errorf("acquire semaphore for %s: %w", providerName, err)
	}
	metrics.LLMConcurrent.WithLabelValues(providerName).Inc()
	defer func() {
		metrics.LLMConcurrent.WithLabelValues(providerName).Dec()
		g.release(providerName)
	}()

	start := time.Now()
	maxRetries := config.C.LLM.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 1
	}

	var lastErr error
	for i := range maxRetries {
		resp, err := p.Chat(ctx, req)
		if err == nil {
			resp.Provider = providerName
			duration := time.Since(start).Seconds()
			metrics.LLMRequestsTotal.WithLabelValues(providerName, req.Model, "success").Inc()
			metrics.LLMRequestDuration.WithLabelValues(providerName, req.Model).Observe(duration)
			metrics.LLMTokensTotal.WithLabelValues(providerName, "input").Add(float64(resp.Usage.PromptTokens))
			metrics.LLMTokensTotal.WithLabelValues(providerName, "output").Add(float64(resp.Usage.CompletionTokens))
			return resp, nil
		}
		lastErr = err

		backoff := retryBackoff(i, err)
		logger.Warn("llm call failed, retrying",
			zap.String("provider", providerName),
			zap.Int("attempt", i+1),
			zap.Duration("backoff", backoff),
			zap.Error(err))

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}
	metrics.LLMRequestsTotal.WithLabelValues(providerName, req.Model, "error").Inc()
	return nil, fmt.Errorf("all retries failed for %s: %w", providerName, lastErr)
}

func (g *Gateway) Embedding(ctx context.Context, req EmbeddingRequest) (*EmbeddingResponse, error) {
	g.mu.RLock()
	providerName := req.Provider
	if providerName == "" {
		providerName = g.fallback
	}

	p, ok := g.providers[providerName]
	if !ok {
		g.mu.RUnlock()
		return nil, fmt.Errorf("provider not available: %s", providerName)
	}

	ep, ok := p.(EmbeddingProvider)
	if !ok {
		g.mu.RUnlock()
		return nil, fmt.Errorf("provider %s does not support embedding", providerName)
	}

	if req.Model == "" {
		req.Model = g.defaults[providerName]
	}
	g.mu.RUnlock()

	if err := g.acquire(ctx, providerName); err != nil {
		return nil, fmt.Errorf("acquire semaphore for %s: %w", providerName, err)
	}
	metrics.LLMConcurrent.WithLabelValues(providerName).Inc()
	defer func() {
		metrics.LLMConcurrent.WithLabelValues(providerName).Dec()
		g.release(providerName)
	}()

	start := time.Now()
	maxRetries := config.C.LLM.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 1
	}

	var lastErr error
	for i := range maxRetries {
		resp, err := ep.Embedding(ctx, req)
		if err == nil {
			resp.Provider = providerName
			duration := time.Since(start).Seconds()
			metrics.LLMRequestsTotal.WithLabelValues(providerName, req.Model, "success").Inc()
			metrics.LLMRequestDuration.WithLabelValues(providerName, req.Model).Observe(duration)
			metrics.LLMTokensTotal.WithLabelValues(providerName, "input").Add(float64(resp.Usage.TotalTokens))
			return resp, nil
		}
		lastErr = err

		backoff := retryBackoff(i, err)
		logger.Warn("embedding call failed, retrying",
			zap.String("provider", providerName),
			zap.Int("attempt", i+1),
			zap.Duration("backoff", backoff),
			zap.Error(err))

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}
	metrics.LLMRequestsTotal.WithLabelValues(providerName, req.Model, "error").Inc()
	return nil, fmt.Errorf("all embedding retries failed for %s: %w", providerName, lastErr)
}

func (g *Gateway) ListModels(ctx context.Context, providerName string) ([]ModelInfo, error) {
	g.mu.RLock()
	p, ok := g.providers[providerName]
	g.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("provider not available: %s", providerName)
	}

	lister, ok := p.(ModelLister)
	if !ok {
		return nil, fmt.Errorf("provider %s does not support listing models", providerName)
	}
	return lister.ListModels(ctx)
}

func (g *Gateway) CalcCost(providerName string, usage Usage) float64 {
	g.mu.RLock()
	p, ok := g.pricing[providerName]
	g.mu.RUnlock()
	if !ok {
		return 0
	}
	inputCost := float64(usage.PromptTokens) / 1_000_000 * p[0]
	outputCost := float64(usage.CompletionTokens) / 1_000_000 * p[1]
	return inputCost + outputCost
}

func (g *Gateway) ChatWithCredential(ctx context.Context, req Request, providerType, apiKey, baseURL string) (*Response, error) {
	var p Provider
	switch providerType {
	case "anthropic":
		p = NewAnthropic(apiKey, baseURL)
	case "doubao":
		p = NewDoubao(apiKey, baseURL)
	default:
		p = NewOpenAI(apiKey, baseURL)
	}

	start := time.Now()
	maxRetries := config.C.LLM.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 1
	}

	var lastErr error
	for i := range maxRetries {
		resp, err := p.Chat(ctx, req)
		if err == nil {
			resp.Provider = providerType
			duration := time.Since(start).Seconds()
			metrics.LLMRequestsTotal.WithLabelValues(providerType, req.Model, "success").Inc()
			metrics.LLMRequestDuration.WithLabelValues(providerType, req.Model).Observe(duration)
			metrics.LLMTokensTotal.WithLabelValues(providerType, "input").Add(float64(resp.Usage.PromptTokens))
			metrics.LLMTokensTotal.WithLabelValues(providerType, "output").Add(float64(resp.Usage.CompletionTokens))
			return resp, nil
		}
		lastErr = err

		backoff := retryBackoff(i, err)
		logger.Warn("llm credential call failed, retrying",
			zap.String("provider", providerType),
			zap.Int("attempt", i+1),
			zap.Duration("backoff", backoff))

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}
	metrics.LLMRequestsTotal.WithLabelValues(providerType, req.Model, "error").Inc()
	return nil, fmt.Errorf("all retries failed for credential provider %s: %w", providerType, lastErr)
}

func (g *Gateway) EmbeddingWithCredential(ctx context.Context, req EmbeddingRequest, providerType, apiKey, baseURL string) (*EmbeddingResponse, error) {
	var p Provider
	switch providerType {
	case "anthropic":
		p = NewAnthropic(apiKey, baseURL)
	case "doubao":
		p = NewDoubao(apiKey, baseURL)
	default:
		p = NewOpenAI(apiKey, baseURL)
	}

	ep, ok := p.(EmbeddingProvider)
	if !ok {
		return nil, fmt.Errorf("provider %s does not support embedding", providerType)
	}

	resp, err := ep.Embedding(ctx, req)
	if err != nil {
		return nil, err
	}
	resp.Provider = providerType
	return resp, nil
}
