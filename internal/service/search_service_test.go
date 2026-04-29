package service

import (
	"context"
	"testing"

	"github.com/mirainya/nexus/pkg/vectordb"
)

type mockVecClient struct {
	results []vectordb.SearchResult
}

func (m *mockVecClient) Insert(collection string, id string, vector []float32, metadata map[string]any) error {
	return nil
}

func (m *mockVecClient) Search(collection string, vector []float32, topK int, filters map[string]any) ([]vectordb.SearchResult, error) {
	return m.results, nil
}

func (m *mockVecClient) Delete(collection string, ids []string) error {
	return nil
}

func TestSearchService_NilVecClient_NoPanic(t *testing.T) {
	svc := NewSearchService(testDB, nil, nil)
	if svc.vec != nil {
		t.Error("expected vec to be nil")
	}
}

func TestSearchService_VectorRecall_NilClient(t *testing.T) {
	svc := NewSearchService(testDB, nil, nil)
	uuids, scores := svc.vectorRecall(context.Background(), "test query", 10)
	if uuids != nil {
		t.Errorf("expected nil uuids, got %v", uuids)
	}
	if scores != nil {
		t.Errorf("expected nil scores, got %v", scores)
	}
}

func TestSearchService_ConstructorWithVec(t *testing.T) {
	mock := &mockVecClient{}
	svc := NewSearchService(testDB, nil, mock)
	if svc.vec == nil {
		t.Error("expected vec to be set")
	}
}

func TestMockVecClient_ImplementsClient(t *testing.T) {
	var _ vectordb.Client = (*mockVecClient)(nil)
}
