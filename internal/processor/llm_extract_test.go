package processor

import "testing"

func TestExtractJSON_PureJSON(t *testing.T) {
	input := `{"entities": [], "relations": []}`
	got := extractJSON(input)
	if got != input {
		t.Errorf("expected %q, got %q", input, got)
	}
}

func TestExtractJSON_MarkdownCodeBlock(t *testing.T) {
	input := "Here is the result:\n```json\n{\"entities\": []}\n```\nDone."
	expected := `{"entities": []}`
	got := extractJSON(input)
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestExtractJSON_GenericCodeBlock(t *testing.T) {
	input := "Result:\n```\n{\"name\": \"test\"}\n```"
	expected := `{"name": "test"}`
	got := extractJSON(input)
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestExtractJSON_SurroundingText(t *testing.T) {
	input := `I found the following: {"entities": [{"name": "Alice"}]} Hope this helps!`
	expected := `{"entities": [{"name": "Alice"}]}`
	got := extractJSON(input)
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestExtractJSON_NoJSON(t *testing.T) {
	input := "No JSON here at all"
	got := extractJSON(input)
	if got != input {
		t.Errorf("expected original string, got %q", got)
	}
}

func TestExtractJSON_NestedBraces(t *testing.T) {
	input := `{"a": {"b": {"c": 1}}}`
	got := extractJSON(input)
	if got != input {
		t.Errorf("expected %q, got %q", input, got)
	}
}

func TestExtractJSON_CodeBlockWithExtraBraces(t *testing.T) {
	input := "Text with {curly} braces\n```json\n{\"real\": \"data\"}\n```\nMore {text}"
	expected := `{"real": "data"}`
	got := extractJSON(input)
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestRenderPrompt(t *testing.T) {
	tmpl := "Extract entities from {{doc_type}} document about {{topic}}"
	vars := map[string]any{"doc_type": "image", "topic": "people"}
	got := renderPrompt(tmpl, vars)
	expected := "Extract entities from image document about people"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestRenderPrompt_Empty(t *testing.T) {
	got := renderPrompt("", nil)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
