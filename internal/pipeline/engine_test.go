package pipeline

import (
	"context"
	"errors"
	"os"
	"sync/atomic"
	"testing"

	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/pkg/config"
	"github.com/mirainya/nexus/pkg/logger"
)

func TestMain(m *testing.M) {
	config.C = &config.Config{}
	logger.Init()
	os.Exit(m.Run())
}

type mockProcessor struct {
	name    string
	err     error
	calls   atomic.Int32
}

func (m *mockProcessor) Name() string { return m.name }
func (m *mockProcessor) Process(_ context.Context, _ *ProcessorContext, _ StepConfig) error {
	m.calls.Add(1)
	return m.err
}

func setupMockProcessor(name string, err error) *mockProcessor {
	p := &mockProcessor{name: name, err: err}
	Register(p)
	return p
}

func TestEngine_SequentialSteps(t *testing.T) {
	p1 := setupMockProcessor("test_seq_1", nil)
	p2 := setupMockProcessor("test_seq_2", nil)
	defer func() {
		mu.Lock()
		delete(registry, "test_seq_1")
		delete(registry, "test_seq_2")
		mu.Unlock()
	}()

	pipe := &model.Pipeline{
		Steps: []model.PipelineStep{
			{SortOrder: 1, ProcessorType: "test_seq_1"},
			{SortOrder: 2, ProcessorType: "test_seq_2"},
		},
	}
	pctx := &ProcessorContext{Document: DocumentData{Type: "text"}}
	engine := NewEngine()

	err := engine.Run(context.Background(), pipe, pctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p1.calls.Load() != 1 {
		t.Errorf("expected p1 called 1 time, got %d", p1.calls.Load())
	}
	if p2.calls.Load() != 1 {
		t.Errorf("expected p2 called 1 time, got %d", p2.calls.Load())
	}
}

func TestEngine_ParallelGroup(t *testing.T) {
	p1 := setupMockProcessor("test_par_1", nil)
	p2 := setupMockProcessor("test_par_2", nil)
	defer func() {
		mu.Lock()
		delete(registry, "test_par_1")
		delete(registry, "test_par_2")
		mu.Unlock()
	}()

	pipe := &model.Pipeline{
		Steps: []model.PipelineStep{
			{SortOrder: 1, ProcessorType: "test_par_1", ParallelGroup: 1},
			{SortOrder: 2, ProcessorType: "test_par_2", ParallelGroup: 1},
		},
	}
	pctx := &ProcessorContext{Document: DocumentData{Type: "text"}}
	engine := NewEngine()

	err := engine.Run(context.Background(), pipe, pctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p1.calls.Load() != 1 {
		t.Errorf("expected p1 called 1 time, got %d", p1.calls.Load())
	}
	if p2.calls.Load() != 1 {
		t.Errorf("expected p2 called 1 time, got %d", p2.calls.Load())
	}
}

func TestEngine_ConditionSkip(t *testing.T) {
	p := setupMockProcessor("test_cond_skip", nil)
	defer func() {
		mu.Lock()
		delete(registry, "test_cond_skip")
		mu.Unlock()
	}()

	pipe := &model.Pipeline{
		Steps: []model.PipelineStep{
			{SortOrder: 1, ProcessorType: "test_cond_skip", Condition: "type=image"},
		},
	}
	pctx := &ProcessorContext{Document: DocumentData{Type: "text"}}
	engine := NewEngine()

	err := engine.Run(context.Background(), pipe, pctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.calls.Load() != 0 {
		t.Errorf("expected processor not called, got %d", p.calls.Load())
	}
}

func TestEngine_ConditionMatch(t *testing.T) {
	p := setupMockProcessor("test_cond_match", nil)
	defer func() {
		mu.Lock()
		delete(registry, "test_cond_match")
		mu.Unlock()
	}()

	pipe := &model.Pipeline{
		Steps: []model.PipelineStep{
			{SortOrder: 1, ProcessorType: "test_cond_match", Condition: "type=image"},
		},
	}
	pctx := &ProcessorContext{Document: DocumentData{Type: "image"}}
	engine := NewEngine()

	err := engine.Run(context.Background(), pipe, pctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.calls.Load() != 1 {
		t.Errorf("expected processor called 1 time, got %d", p.calls.Load())
	}
}

func TestEngine_ErrorStop(t *testing.T) {
	setupMockProcessor("test_err_stop", errors.New("boom"))
	p2 := setupMockProcessor("test_err_stop_2", nil)
	defer func() {
		mu.Lock()
		delete(registry, "test_err_stop")
		delete(registry, "test_err_stop_2")
		mu.Unlock()
	}()

	pipe := &model.Pipeline{
		Steps: []model.PipelineStep{
			{SortOrder: 1, ProcessorType: "test_err_stop", OnError: "stop"},
			{SortOrder: 2, ProcessorType: "test_err_stop_2"},
		},
	}
	pctx := &ProcessorContext{Document: DocumentData{Type: "text"}}
	engine := NewEngine()

	err := engine.Run(context.Background(), pipe, pctx)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if p2.calls.Load() != 0 {
		t.Errorf("expected p2 not called after stop, got %d", p2.calls.Load())
	}
}

func TestEngine_ErrorSkip(t *testing.T) {
	setupMockProcessor("test_err_skip", errors.New("boom"))
	p2 := setupMockProcessor("test_err_skip_2", nil)
	defer func() {
		mu.Lock()
		delete(registry, "test_err_skip")
		delete(registry, "test_err_skip_2")
		mu.Unlock()
	}()

	pipe := &model.Pipeline{
		Steps: []model.PipelineStep{
			{SortOrder: 1, ProcessorType: "test_err_skip", OnError: "skip"},
			{SortOrder: 2, ProcessorType: "test_err_skip_2"},
		},
	}
	pctx := &ProcessorContext{Document: DocumentData{Type: "text"}}
	engine := NewEngine()

	err := engine.Run(context.Background(), pipe, pctx)
	if !errors.Is(err, ErrPartial) {
		t.Fatalf("expected ErrPartial, got %v", err)
	}
	if p2.calls.Load() != 1 {
		t.Errorf("expected p2 called after skip, got %d", p2.calls.Load())
	}
}

func TestEngine_ErrorRetry(t *testing.T) {
	callCount := atomic.Int32{}
	retryProc := &retryMockProcessor{calls: &callCount, failUntil: 2}
	mu.Lock()
	registry["test_err_retry"] = retryProc
	mu.Unlock()
	defer func() {
		mu.Lock()
		delete(registry, "test_err_retry")
		mu.Unlock()
	}()

	pipe := &model.Pipeline{
		Steps: []model.PipelineStep{
			{SortOrder: 1, ProcessorType: "test_err_retry", OnError: "retry", MaxRetry: 3},
		},
	}
	pctx := &ProcessorContext{Document: DocumentData{Type: "text"}}
	engine := NewEngine()

	err := engine.Run(context.Background(), pipe, pctx)
	if err != nil {
		t.Fatalf("expected success after retry, got %v", err)
	}
	if callCount.Load() != 3 {
		t.Errorf("expected 3 calls (2 fail + 1 success), got %d", callCount.Load())
	}
}

type retryMockProcessor struct {
	calls     *atomic.Int32
	failUntil int32
}

func (m *retryMockProcessor) Name() string { return "test_err_retry" }
func (m *retryMockProcessor) Process(_ context.Context, _ *ProcessorContext, _ StepConfig) error {
	n := m.calls.Add(1)
	if n <= m.failUntil {
		return errors.New("transient error")
	}
	return nil
}

func TestEngine_StartFrom(t *testing.T) {
	p1 := setupMockProcessor("test_sf_1", nil)
	p2 := setupMockProcessor("test_sf_2", nil)
	defer func() {
		mu.Lock()
		delete(registry, "test_sf_1")
		delete(registry, "test_sf_2")
		mu.Unlock()
	}()

	pipe := &model.Pipeline{
		Steps: []model.PipelineStep{
			{SortOrder: 1, ProcessorType: "test_sf_1"},
			{SortOrder: 2, ProcessorType: "test_sf_2"},
		},
	}
	pctx := &ProcessorContext{Document: DocumentData{Type: "text"}}
	engine := NewEngine()

	err := engine.Run(context.Background(), pipe, pctx, WithStartFrom(2))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p1.calls.Load() != 0 {
		t.Errorf("expected p1 skipped, got %d calls", p1.calls.Load())
	}
	if p2.calls.Load() != 1 {
		t.Errorf("expected p2 called 1 time, got %d", p2.calls.Load())
	}
}

func TestEvalCondition_TypeEquals(t *testing.T) {
	pctx := &ProcessorContext{Document: DocumentData{Type: "image"}}
	if !EvalCondition("type=image", pctx) {
		t.Error("expected true for type=image")
	}
	if EvalCondition("type=text", pctx) {
		t.Error("expected false for type=text")
	}
}

func TestEvalCondition_HasEntities(t *testing.T) {
	pctx := &ProcessorContext{}
	if EvalCondition("has:entities", pctx) {
		t.Error("expected false when no entities")
	}
	pctx.Entities = []EntityData{{Name: "test"}}
	if !EvalCondition("has:entities", pctx) {
		t.Error("expected true when entities exist")
	}
}

func TestEvalCondition_Contains(t *testing.T) {
	pctx := &ProcessorContext{
		Extras: map[string]any{
			"classification": map[string]any{
				"tags": "photo,portrait",
			},
		},
	}
	if !EvalCondition("classification.tags contains photo", pctx) {
		t.Error("expected true for contains photo")
	}
	if EvalCondition("classification.tags contains video", pctx) {
		t.Error("expected false for contains video")
	}
}

func TestEvalCondition_DotNotation(t *testing.T) {
	pctx := &ProcessorContext{
		Extras: map[string]any{
			"classification": map[string]any{
				"category": "person",
			},
		},
	}
	if !EvalCondition("classification.category=person", pctx) {
		t.Error("expected true for classification.category=person")
	}
	if EvalCondition("classification.category=animal", pctx) {
		t.Error("expected false for classification.category=animal")
	}
}

func TestEvalCondition_Empty(t *testing.T) {
	pctx := &ProcessorContext{}
	if !EvalCondition("", pctx) {
		t.Error("expected true for empty condition")
	}
}

func TestEngine_Callbacks(t *testing.T) {
	setupMockProcessor("test_cb", nil)
	defer func() {
		mu.Lock()
		delete(registry, "test_cb")
		mu.Unlock()
	}()

	pipe := &model.Pipeline{
		Steps: []model.PipelineStep{
			{SortOrder: 1, ProcessorType: "test_cb"},
		},
	}
	pctx := &ProcessorContext{Document: DocumentData{Type: "text"}}
	engine := NewEngine()

	var startCalled, endCalled, progressCalled bool
	err := engine.Run(context.Background(), pipe, pctx,
		WithOnStepStart(func(order int, pt string) { startCalled = true }),
		WithOnStepEnd(func(order int, pt string, err error, log StepLog) { endCalled = true }),
		WithOnProgress(func(order int) { progressCalled = true }),
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !startCalled {
		t.Error("OnStepStart not called")
	}
	if !endCalled {
		t.Error("OnStepEnd not called")
	}
	if !progressCalled {
		t.Error("OnProgress not called")
	}
}
