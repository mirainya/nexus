package pipeline

import "context"

// ProcessorContext is the data passed between pipeline steps.
type ProcessorContext struct {
	Document       DocumentData           `json:"document"`
	Entities       []EntityData           `json:"entities"`
	Relations      []RelationData         `json:"relations"`
	RawText        string                 `json:"raw_text"`
	Summary        string                 `json:"summary"`
	Extras         map[string]any         `json:"extras"`
	StepLogs       []StepLog              `json:"step_logs"`
	SourceImageURL string                 `json:"-"`
	ImageBase64    string                 `json:"-"`
}

type DocumentData struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Content   string         `json:"content"`
	SourceURL string         `json:"source_url"`
	Metadata  map[string]any `json:"metadata"`
}

type EntityData struct {
	Type       string         `json:"type"`
	Name       string         `json:"name"`
	Aliases    []string       `json:"aliases,omitempty"`
	Confidence float64        `json:"confidence"`
	Attributes map[string]any `json:"attributes,omitempty"`
	Evidence   map[string]any `json:"evidence,omitempty"`
}

type RelationData struct {
	From       string         `json:"from"`
	To         string         `json:"to"`
	Type       string         `json:"type"`
	Confidence float64        `json:"confidence"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type StepLog struct {
	Processor string  `json:"processor"`
	Duration  float64 `json:"duration_ms"`
	Tokens    int     `json:"tokens,omitempty"`
	Cost      float64 `json:"cost,omitempty"`
	Error     string  `json:"error,omitempty"`
}

// StepConfig holds the configuration for a single pipeline step.
type StepConfig struct {
	ProcessorType    string         `json:"processor_type"`
	PromptContent    string         `json:"prompt_content"`
	PromptVariables  map[string]any `json:"prompt_variables"`
	Config           map[string]any `json:"config"`
	Condition        string         `json:"condition"`
}

// Processor is the interface all processors must implement.
type Processor interface {
	Name() string
	Process(ctx context.Context, pctx *ProcessorContext, cfg StepConfig) error
}

type RunOptions struct {
	StartFrom   int
	OnProgress  func(stepOrder int)
	OnStepStart func(stepOrder int, processorType string)
	OnStepEnd   func(stepOrder int, processorType string, err error, log StepLog)
}

type RunOption func(*RunOptions)

func WithStartFrom(step int) RunOption {
	return func(o *RunOptions) { o.StartFrom = step }
}

func WithOnProgress(fn func(stepOrder int)) RunOption {
	return func(o *RunOptions) { o.OnProgress = fn }
}

func WithOnStepStart(fn func(stepOrder int, processorType string)) RunOption {
	return func(o *RunOptions) { o.OnStepStart = fn }
}

func WithOnStepEnd(fn func(stepOrder int, processorType string, err error, log StepLog)) RunOption {
	return func(o *RunOptions) { o.OnStepEnd = fn }
}
