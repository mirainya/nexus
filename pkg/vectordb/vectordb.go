package vectordb

// Client defines the interface for vector database operations.
type Client interface {
	Insert(collection string, id string, vector []float32, metadata map[string]any) error
	Search(collection string, vector []float32, topK int, filters map[string]any) ([]SearchResult, error)
	Delete(collection string, ids []string) error
}

type SearchResult struct {
	ID       string         `json:"id"`
	Score    float32        `json:"score"`
	Metadata map[string]any `json:"metadata"`
}
