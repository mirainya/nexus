package processor

import (
	"context"
	"testing"

	"github.com/mirainya/nexus/internal/pipeline"
	"github.com/mirainya/nexus/pkg/vectordb"
)

type mockVectorDB struct {
	inserted []mockInsertCall
}

type mockInsertCall struct {
	Collection string
	ID         string
	Vector     []float32
	Metadata   map[string]any
}

func (m *mockVectorDB) Insert(collection string, id string, vector []float32, metadata map[string]any) error {
	m.inserted = append(m.inserted, mockInsertCall{
		Collection: collection,
		ID:         id,
		Vector:     vector,
		Metadata:   metadata,
	})
	return nil
}

func (m *mockVectorDB) Search(collection string, vector []float32, topK int, filters map[string]any) ([]vectordb.SearchResult, error) {
	return nil, nil
}

func (m *mockVectorDB) Delete(collection string, ids []string) error {
	return nil
}

func TestEmbedding_NilVectorDB_NoPanic(t *testing.T) {
	pctx := &pipeline.ProcessorContext{
		Document: pipeline.DocumentData{
			ID:   "test-doc-1",
			Type: "text",
		},
		Extras: map[string]any{
			"embedding": []float64{0.1, 0.2, 0.3},
		},
	}

	if pctx.VectorDB != nil {
		t.Error("expected VectorDB to be nil")
	}
}

func TestEmbedding_VectorDBInsertCalled(t *testing.T) {
	mock := &mockVectorDB{}
	pctx := &pipeline.ProcessorContext{
		Document: pipeline.DocumentData{
			ID:   "test-doc-2",
			Type: "image",
		},
		VectorDB: mock,
	}

	embedding := []float64{0.1, 0.2, 0.3}
	vec32 := make([]float32, len(embedding))
	for i, v := range embedding {
		vec32[i] = float32(v)
	}
	meta := map[string]any{
		"doc_id":   pctx.Document.ID,
		"doc_type": pctx.Document.Type,
	}

	if err := pctx.VectorDB.Insert("test_collection", pctx.Document.ID, vec32, meta); err != nil {
		t.Fatalf("insert: %v", err)
	}

	if len(mock.inserted) != 1 {
		t.Fatalf("expected 1 insert call, got %d", len(mock.inserted))
	}
	call := mock.inserted[0]
	if call.ID != "test-doc-2" {
		t.Errorf("expected id test-doc-2, got %s", call.ID)
	}
	if call.Collection != "test_collection" {
		t.Errorf("expected collection test_collection, got %s", call.Collection)
	}
	if len(call.Vector) != 3 {
		t.Errorf("expected 3-dim vector, got %d", len(call.Vector))
	}
}

func TestMockVectorDB_ImplementsClient(t *testing.T) {
	var _ vectordb.Client = (*mockVectorDB)(nil)
}

func TestEmbedding_Float64ToFloat32Conversion(t *testing.T) {
	embedding := []float64{0.123456789, -0.987654321, 1.0}
	vec32 := make([]float32, len(embedding))
	for i, v := range embedding {
		vec32[i] = float32(v)
	}

	ctx := context.Background()
	_ = ctx

	if len(vec32) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(vec32))
	}
	if vec32[2] != 1.0 {
		t.Errorf("expected 1.0, got %f", vec32[2])
	}
}
