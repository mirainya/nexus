package vectordb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/mirainya/nexus/pkg/logger"
	"go.uber.org/zap"
)

var (
	defaultClient  Client
	defaultCollection string
)

func Init(addr, collection string, dim int) error {
	if addr == "" {
		logger.Info("milvus not configured, vector search disabled")
		return nil
	}
	if collection == "" {
		collection = "nexus_embeddings"
	}
	if dim <= 0 {
		dim = 1536
	}

	mc, err := NewMilvusClient(addr, collection, dim)
	if err != nil {
		return err
	}
	defaultClient = mc
	defaultCollection = collection
	return nil
}

func Available() bool { return defaultClient != nil }

func Default() Client { return defaultClient }

func Collection() string { return defaultCollection }

func Close() {
	if mc, ok := defaultClient.(*MilvusClient); ok && mc != nil {
		mc.Close()
	}
}

type MilvusClient struct {
	mc  client.Client
	dim int
}

func NewMilvusClient(addr, collection string, dim int) (*MilvusClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mc, err := client.NewDefaultGrpcClient(ctx, addr)
	if err != nil {
		return nil, fmt.Errorf("milvus connect: %w", err)
	}

	m := &MilvusClient{mc: mc, dim: dim}
	if err := m.ensureCollection(ctx, collection); err != nil {
		mc.Close()
		return nil, err
	}

	logger.Info("milvus connected", zap.String("addr", addr), zap.String("collection", collection))
	return m, nil
}

func (m *MilvusClient) ensureCollection(ctx context.Context, collection string) error {
	has, err := m.mc.HasCollection(ctx, collection)
	if err != nil {
		return fmt.Errorf("milvus has collection: %w", err)
	}
	if has {
		if err := m.mc.LoadCollection(ctx, collection, true); err != nil {
			logger.Warn("milvus load collection", zap.Error(err))
		}
		return nil
	}

	schema := entity.NewSchema().
		WithName(collection).
		WithField(entity.NewField().WithName("id").WithDataType(entity.FieldTypeVarChar).WithMaxLength(128).WithIsPrimaryKey(true)).
		WithField(entity.NewField().WithName("vector").WithDataType(entity.FieldTypeFloatVector).WithDim(int64(m.dim))).
		WithField(entity.NewField().WithName("metadata").WithDataType(entity.FieldTypeJSON))

	if err := m.mc.CreateCollection(ctx, schema, 1); err != nil {
		return fmt.Errorf("milvus create collection: %w", err)
	}

	idx, err := entity.NewIndexIvfFlat(entity.IP, 128)
	if err != nil {
		return fmt.Errorf("milvus create index params: %w", err)
	}
	if err := m.mc.CreateIndex(ctx, collection, "vector", idx, false); err != nil {
		return fmt.Errorf("milvus create index: %w", err)
	}

	if err := m.mc.LoadCollection(ctx, collection, true); err != nil {
		return fmt.Errorf("milvus load collection: %w", err)
	}

	return nil
}

func (m *MilvusClient) Insert(collection string, id string, vector []float32, metadata map[string]any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	metaBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	idCol := entity.NewColumnVarChar("id", []string{id})
	vecCol := entity.NewColumnFloatVector("vector", m.dim, [][]float32{vector})
	metaCol := entity.NewColumnJSONBytes("metadata", [][]byte{metaBytes})

	_, err = m.mc.Upsert(ctx, collection, "", idCol, vecCol, metaCol)
	if err != nil {
		return fmt.Errorf("milvus upsert: %w", err)
	}
	return nil
}

func (m *MilvusClient) Search(collection string, vector []float32, topK int, filters map[string]any) ([]SearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sp, err := entity.NewIndexIvfFlatSearchParam(16)
	if err != nil {
		return nil, fmt.Errorf("milvus search param: %w", err)
	}

	expr := ""
	vectors := []entity.Vector{entity.FloatVector(vector)}

	results, err := m.mc.Search(ctx, collection, nil, expr, []string{"id", "metadata"}, vectors, "vector", entity.IP, topK, sp)
	if err != nil {
		return nil, fmt.Errorf("milvus search: %w", err)
	}

	var out []SearchResult
	for _, result := range results {
		for i := 0; i < result.ResultCount; i++ {
			sr := SearchResult{Score: result.Scores[i]}

			if idCol, ok := result.IDs.(*entity.ColumnVarChar); ok {
				val, err := idCol.ValueByIdx(i)
				if err == nil {
					sr.ID = val
				}
			}

			for _, field := range result.Fields {
				if field.Name() == "metadata" {
					if jsonCol, ok := field.(*entity.ColumnJSONBytes); ok {
						raw, err := jsonCol.ValueByIdx(i)
						if err == nil {
							var meta map[string]any
							if json.Unmarshal(raw, &meta) == nil {
								sr.Metadata = meta
							}
						}
					}
				}
			}

			out = append(out, sr)
		}
	}
	return out, nil
}

func (m *MilvusClient) Delete(collection string, ids []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	idCol := entity.NewColumnVarChar("id", ids)
	return m.mc.DeleteByPks(ctx, collection, "", idCol)
}

func (m *MilvusClient) Close() error {
	return m.mc.Close()
}
