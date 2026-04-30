package service

import (
	"github.com/mirainya/nexus/internal/model"
	"gorm.io/gorm"
)

type GraphService struct{ db *gorm.DB }

func NewGraphService(db *gorm.DB) *GraphService { return &GraphService{db: db} }

type GraphNode struct {
	ID         uint   `json:"id"`
	Label      string `json:"label"`
	Type       string `json:"type"`
	Confidence float64 `json:"confidence"`
}

type GraphEdge struct {
	Source uint   `json:"source"`
	Target uint   `json:"target"`
	Type   string `json:"type"`
}

type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

func (s *GraphService) GetGraphData(limit int, tenantID uint) (*GraphData, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}

	var entities []model.Entity
	q := s.db.Limit(limit)
	if tenantID > 0 {
		q = q.Where("tenant_id = ?", tenantID)
	}
	if err := q.Find(&entities).Error; err != nil {
		return nil, err
	}

	idSet := make(map[uint]bool, len(entities))
	nodes := make([]GraphNode, 0, len(entities))
	for _, e := range entities {
		idSet[e.ID] = true
		nodes = append(nodes, GraphNode{
			ID:         e.ID,
			Label:      e.Name,
			Type:       e.Type,
			Confidence: e.Confidence,
		})
	}

	var relations []model.Relation
	rq := s.db.Model(&model.Relation{}).Limit(limit * 5)
	if tenantID > 0 {
		rq = rq.Where("tenant_id = ?", tenantID)
	}
	if err := rq.Find(&relations).Error; err != nil {
		return nil, err
	}

	edges := make([]GraphEdge, 0)
	for _, r := range relations {
		if idSet[r.FromEntityID] && idSet[r.ToEntityID] {
			edges = append(edges, GraphEdge{
				Source: r.FromEntityID,
				Target: r.ToEntityID,
				Type:   r.Type,
			})
		}
	}

	return &GraphData{Nodes: nodes, Edges: edges}, nil
}
