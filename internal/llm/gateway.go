package llm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/pkg/config"
	"github.com/mirainya/nexus/pkg/logger"
	"go.uber.org/zap"
)

type Gateway struct {
	providers map[string]Provider
	defaults  map[string]string
	pricing   map[string][2]float64 // [input_price, output_price] per 1M tokens
	fallback  string
	mu        sync.RWMutex
}

var G *Gateway

func Init() {
	G = &Gateway{
		providers: make(map[string]Provider),
		defaults:  make(map[string]string),
		pricing:   make(map[string][2]float64),
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
		if p.IsDefault {
			g.fallback = p.Name
		}
		logger.Info("llm provider registered", zap.String("provider", p.Name))
	}

	if g.fallback == "" && len(g.providers) > 0 {
		for name := range g.providers {
			g.fallback = name
			break
		}
	}
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

	maxRetries := config.C.LLM.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 1
	}

	var lastErr error
	for i := range maxRetries {
		resp, err := p.Chat(ctx, req)
		if err == nil {
			resp.Provider = providerName
			return resp, nil
		}
		lastErr = err
		logger.Warn("llm call failed, retrying",
			zap.String("provider", providerName),
			zap.Int("attempt", i+1),
			zap.Error(err))
		time.Sleep(time.Duration(i+1) * time.Second)
	}
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

	maxRetries := config.C.LLM.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 1
	}

	var lastErr error
	for i := range maxRetries {
		resp, err := ep.Embedding(ctx, req)
		if err == nil {
			resp.Provider = providerName
			return resp, nil
		}
		lastErr = err
		logger.Warn("embedding call failed, retrying",
			zap.String("provider", providerName),
			zap.Int("attempt", i+1),
			zap.Error(err))
		time.Sleep(time.Duration(i+1) * time.Second)
	}
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
