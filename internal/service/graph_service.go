package service

import "github.com/mirainya/nexus/internal/model"

type GraphService struct{}

func NewGraphService() *GraphService { return &GraphService{} }

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

func (s *GraphService) GetGraphData(limit int) (*GraphData, error) {
	db := model.DB()
	if limit <= 0 || limit > 500 {
		limit = 200
	}

	var entities []model.Entity
	if err := db.Limit(limit).Find(&entities).Error; err != nil {
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
	if err := db.Find(&relations).Error; err != nil {
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
