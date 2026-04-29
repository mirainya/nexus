package vectordb

import "testing"

var _ Client = (*MilvusClient)(nil)

func TestMilvusClient_ImplementsClient(t *testing.T) {
	// compile-time check above is sufficient
}
