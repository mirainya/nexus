package llm

import (
	"context"
	"errors"
	"math"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mirainya/nexus/pkg/config"
	"github.com/mirainya/nexus/pkg/logger"
)

func TestMain(m *testing.M) {
	config.C = &config.Config{LLM: config.LLMConfig{MaxRetries: 1}}
	logger.Init()
	os.Exit(m.Run())
}

type mockProvider struct {
	name     string
	resp     *Response
	err      error
	calls    atomic.Int32
	delay    time.Duration
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Chat(_ context.Context, req Request) (*Response, error) {
	m.calls.Add(1)
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.err != nil {
		return nil, m.err
	}
	resp := *m.resp
	resp.Model = req.Model
	return &resp, nil
}

func newTestGateway() *Gateway {
	return &Gateway{
		providers:  make(map[string]Provider),
		defaults:   make(map[string]string),
		pricing:    make(map[string][2]float64),
		semaphores: make(map[string]chan struct{}),
	}
}

func TestGateway_Chat_Success(t *testing.T) {
	g := newTestGateway()
	mp := &mockProvider{name: "test", resp: &Response{Content: "hello", Usage: Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15}}}
	g.providers["test"] = mp
	g.defaults["test"] = "model-1"
	g.semaphores["test"] = make(chan struct{}, 10)
	g.fallback = "test"

	resp, err := g.Chat(context.Background(), Request{Messages: []Message{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Content != "hello" {
		t.Errorf("expected content 'hello', got %q", resp.Content)
	}
	if resp.Provider != "test" {
		t.Errorf("expected provider 'test', got %q", resp.Provider)
	}
	if resp.Model != "model-1" {
		t.Errorf("expected model 'model-1', got %q", resp.Model)
	}
}

func TestGateway_Chat_ProviderNotFound(t *testing.T) {
	g := newTestGateway()
	_, err := g.Chat(context.Background(), Request{Provider: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing provider")
	}
}

func TestGateway_Chat_FallbackProvider(t *testing.T) {
	g := newTestGateway()
	mp := &mockProvider{name: "fallback", resp: &Response{Content: "ok"}}
	g.providers["fallback"] = mp
	g.defaults["fallback"] = "m1"
	g.semaphores["fallback"] = make(chan struct{}, 10)
	g.fallback = "fallback"

	resp, err := g.Chat(context.Background(), Request{Messages: []Message{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Provider != "fallback" {
		t.Errorf("expected fallback provider, got %q", resp.Provider)
	}
}

func TestGateway_Semaphore(t *testing.T) {
	g := newTestGateway()
	mp := &mockProvider{name: "sem", resp: &Response{Content: "ok"}, delay: 50 * time.Millisecond}
	g.providers["sem"] = mp
	g.defaults["sem"] = "m1"
	g.semaphores["sem"] = make(chan struct{}, 1)
	g.fallback = "sem"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Fill the semaphore
	g.semaphores["sem"] <- struct{}{}

	_, err := g.Chat(ctx, Request{Messages: []Message{{Role: "user", Content: "hi"}}})
	if err == nil {
		t.Fatal("expected context deadline error")
	}
}

func TestGateway_CalcCost(t *testing.T) {
	g := newTestGateway()
	g.pricing["openai"] = [2]float64{3.0, 15.0} // $3/M input, $15/M output

	usage := Usage{PromptTokens: 1000, CompletionTokens: 500}
	cost := g.CalcCost("openai", usage)

	expected := (1000.0/1_000_000)*3.0 + (500.0/1_000_000)*15.0
	if math.Abs(cost-expected) > 1e-9 {
		t.Errorf("expected cost %f, got %f", expected, cost)
	}
}

func TestGateway_CalcCost_UnknownProvider(t *testing.T) {
	g := newTestGateway()
	cost := g.CalcCost("unknown", Usage{PromptTokens: 100})
	if cost != 0 {
		t.Errorf("expected 0 for unknown provider, got %f", cost)
	}
}

func TestRetryBackoff_RateLimit(t *testing.T) {
	err := &RateLimitError{Provider: "test", RetryAfter: 5}
	d := retryBackoff(0, err)
	if d != 5*time.Second {
		t.Errorf("expected 5s for rate limit, got %v", d)
	}
}

func TestRetryBackoff_Exponential(t *testing.T) {
	err := errors.New("generic error")
	d0 := retryBackoff(0, err)
	d1 := retryBackoff(1, err)
	d2 := retryBackoff(2, err)

	if d0 != 1*time.Second {
		t.Errorf("attempt 0: expected 1s, got %v", d0)
	}
	if d1 != 2*time.Second {
		t.Errorf("attempt 1: expected 2s, got %v", d1)
	}
	if d2 != 4*time.Second {
		t.Errorf("attempt 2: expected 4s, got %v", d2)
	}
}

func TestRetryBackoff_MaxCap(t *testing.T) {
	err := errors.New("generic error")
	d := retryBackoff(10, err)
	if d != 30*time.Second {
		t.Errorf("expected 30s cap, got %v", d)
	}
}
