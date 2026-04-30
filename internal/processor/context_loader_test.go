package processor

import "testing"

func TestNameInContent(t *testing.T) {
	tests := []struct {
		name, content string
		want          bool
	}{
		{"Alice", "alice went to the store", true},
		{"ALICE", "alice went to the store", true},
		{"Bob", "alice went to the store", false},
		{"", "anything", false},
		{"test", "", false},
	}
	for _, tt := range tests {
		if got := nameInContent(tt.name, tt.content); got != tt.want {
			t.Errorf("nameInContent(%q, %q) = %v, want %v", tt.name, tt.content, got, tt.want)
		}
	}
}

func TestAliasInContent(t *testing.T) {
	tests := []struct {
		aliases string
		content string
		want    bool
	}{
		{`["Bob","Robert"]`, "robert is here", true},
		{`["Bob","Robert"]`, "alice is here", false},
		{`[]`, "anything", false},
		{``, "anything", false},
		{`invalid`, "anything", false},
	}
	for _, tt := range tests {
		if got := aliasInContent([]byte(tt.aliases), tt.content); got != tt.want {
			t.Errorf("aliasInContent(%s, %q) = %v, want %v", tt.aliases, tt.content, got, tt.want)
		}
	}
}
