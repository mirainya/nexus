package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mirainya/nexus/internal/llm"
	"github.com/mirainya/nexus/internal/model"
	"github.com/mirainya/nexus/pkg/config"
	"github.com/mirainya/nexus/pkg/logger"
	"github.com/mirainya/nexus/pkg/vectordb"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SearchService struct {
	db  *gorm.DB
	gw  *llm.Gateway
	vec vectordb.Client
}

func NewSearchService(db *gorm.DB, gw *llm.Gateway, vec vectordb.Client) *SearchService {
	return &SearchService{db: db, gw: gw, vec: vec}
}

type SearchRequest struct {
	Query    string `json:"query" binding:"required"`
	Limit    int    `json:"limit"`
	Mode     string `json:"mode"`
	TenantID uint   `json:"-"`
}

type SearchStats struct {
	Total       int `json:"total"`
	VectorHits  int `json:"vector_hits"`
	KeywordHits int `json:"keyword_hits"`
}

type SearchResult struct {
	Items     []SearchItem `json:"items"`
	Query     ParsedQuery  `json:"parsed_query"`
	Reasoning string       `json:"reasoning,omitempty"`
	Stats     SearchStats  `json:"stats"`
}

type SearchItem struct {
	DocumentID   uint          `json:"document_id"`
	DocumentUUID string        `json:"document_uuid,omitempty"`
	Type         string        `json:"type"`
	SourceURL    string        `json:"source_url"`
	Content      string        `json:"content,omitempty"`
	Summary      string        `json:"summary,omitempty"`
	Entities     []EntityBrief `json:"entities,omitempty"`
	Score        float64       `json:"score,omitempty"`
	VectorScore  float64       `json:"vector_score,omitempty"`
	Reason       string        `json:"reason,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
}

type EntityBrief struct {
	ID   uint   `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type ParsedQuery struct {
	Entity   string   `json:"entity,omitempty"`
	Type     string   `json:"type,omitempty"`
	DateFrom string   `json:"date_from,omitempty"`
	DateTo   string   `json:"date_to,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
	Intent   string   `json:"intent,omitempty"`
}

func (s *SearchService) Search(ctx context.Context, req SearchRequest) (*SearchResult, error) {
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	mode := req.Mode
	if mode == "" {
		mode = "hybrid"
	}

	result := &SearchResult{}

	switch mode {
	case "vector":
		vecUUIDs, vecScores := s.vectorRecall(ctx, req.Query, limit)
		if len(vecUUIDs) == 0 {
			return result, nil
		}
		items, err := s.queryDocuments(ParsedQuery{}, limit, vecUUIDs, req.TenantID)
		if err != nil {
			return nil, err
		}
		for i := range items {
			if score, ok := vecScores[items[i].DocumentUUID]; ok {
				items[i].VectorScore = float64(score)
			}
		}
		result.Items = items
		result.Stats = SearchStats{Total: len(items), VectorHits: len(items)}

	case "keyword":
		parsed := s.parseIntent(ctx, req.Query)
		result.Query = parsed
		items, err := s.queryDocuments(parsed, limit, nil, req.TenantID)
		if err != nil {
			return nil, err
		}
		result.Items = items
		result.Stats = SearchStats{Total: len(items), KeywordHits: len(items)}
		if len(items) > 5 && parsed.Intent != "" {
			ranked, reasoning := s.rerank(ctx, req.Query, parsed, items, limit)
			if ranked != nil {
				result.Items = ranked
				result.Reasoning = reasoning
				result.Stats.Total = len(ranked)
			}
		}

	default: // hybrid
		parsed := s.parseIntent(ctx, req.Query)
		result.Query = parsed
		vecUUIDs, vecScores := s.vectorRecall(ctx, req.Query, limit)
		items, err := s.queryDocuments(parsed, limit, vecUUIDs, req.TenantID)
		if err != nil {
			return nil, err
		}

		vecSet := make(map[string]bool, len(vecUUIDs))
		for _, u := range vecUUIDs {
			vecSet[u] = true
		}

		vectorHits, keywordHits := 0, 0
		for i := range items {
			isVec := vecSet[items[i].DocumentUUID]
			if isVec {
				if score, ok := vecScores[items[i].DocumentUUID]; ok {
					items[i].VectorScore = float64(score)
				}
				vectorHits++
			} else {
				keywordHits++
			}
		}

		result.Items = items
		result.Stats = SearchStats{Total: len(items), VectorHits: vectorHits, KeywordHits: keywordHits}

		if len(items) > 5 && parsed.Intent != "" {
			ranked, reasoning := s.rerank(ctx, req.Query, parsed, items, limit)
			if ranked != nil {
				result.Items = ranked
				result.Reasoning = reasoning
				result.Stats.Total = len(ranked)
			}
		}
	}

	return result, nil
}

// --- 向量召回 ---

func (s *SearchService) vectorRecall(ctx context.Context, query string, topK int) ([]string, map[string]float32) {
	if s.vec == nil {
		return nil, nil
	}

	resp, err := s.gw.Embedding(ctx, llm.EmbeddingRequest{
		Model: "text-embedding-3-small",
		Input: query,
	})
	if err != nil {
		logger.Warn("vector recall embedding failed", zap.Error(err))
		return nil, nil
	}

	vec32 := make([]float32, len(resp.Embedding))
	for i, v := range resp.Embedding {
		vec32[i] = float32(v)
	}

	collection := config.C.Milvus.Collection
	if collection == "" {
		collection = "nexus_embeddings"
	}

	results, err := s.vec.Search(collection, vec32, topK, nil)
	if err != nil {
		logger.Warn("vector recall search failed", zap.Error(err))
		return nil, nil
	}

	uuids := make([]string, 0, len(results))
	scores := make(map[string]float32, len(results))
	for _, r := range results {
		uuids = append(uuids, r.ID)
		scores[r.ID] = r.Score
	}
	return uuids, scores
}

// --- 阶段 A：LLM 意图解析 ---

const intentSystemPrompt = `你是一个查询意图解析器。将用户的自然语言查询转为结构化 JSON。

输出格式：
{
  "entity": "实体名（人名、组织名等），没有则留空",
  "type": "文档类型：text 或 image，不确定则留空",
  "date_from": "YYYY-MM-DD 格式，没有则留空",
  "date_to": "YYYY-MM-DD 格式，没有则留空",
  "keywords": ["其他关键词"],
  "intent": "用户的深层意图描述，用于后续精排"
}

注意：
- 如果用户提到某个日期但没有范围，date_from 和 date_to 设为同一天
- "照片"、"图片" 对应 type: "image"
- "报告"、"文档"、"文本" 对应 type: "text"
- 当前日期：%s
- 只输出 JSON，不要其他内容`

func (s *SearchService) parseIntent(ctx context.Context, query string) ParsedQuery {
	today := time.Now().Format("2006-01-02")
	resp, err := s.gw.Chat(ctx, llm.Request{
		Messages: []llm.Message{
			{Role: "system", Content: fmt.Sprintf(intentSystemPrompt, today)},
			{Role: "user", Content: query},
		},
		Temperature: 0,
	})
	if err != nil {
		logger.Warn("search intent parse failed, fallback to keyword", zap.Error(err))
		return ParsedQuery{Keywords: strings.Fields(query)}
	}

	var parsed ParsedQuery
	raw := resp.Content
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		raw = raw[start : end+1]
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		logger.Warn("search intent JSON parse failed", zap.String("raw", resp.Content), zap.Error(err))
		return ParsedQuery{Keywords: strings.Fields(query)}
	}
	return parsed
}

// --- 阶段 B：联合查询 ---

type docRow struct {
	model.Document
	JobResult json.RawMessage `gorm:"column:job_result"`
}

func (s *SearchService) queryDocuments(pq ParsedQuery, limit int, vecUUIDs []string, tenantID uint) ([]SearchItem, error) {
	q := s.db.Table("documents d").
		Select("d.*, j.result as job_result").
		Joins("LEFT JOIN jobs j ON j.document_id = d.id AND j.status = 'completed'").
		Where("d.deleted_at IS NULL")

	if tenantID > 0 {
		q = q.Where("d.tenant_id = ?", tenantID)
	}

	hasCondition := false

	var keywordScope *gorm.DB
	keywordScope = s.db.Table("documents d").
		Select("d.*, j.result as job_result").
		Joins("LEFT JOIN jobs j ON j.document_id = d.id AND j.status = 'completed'").
		Where("d.deleted_at IS NULL")

	if tenantID > 0 {
		keywordScope = keywordScope.Where("d.tenant_id = ?", tenantID)
	}

	if pq.Entity != "" {
		keywordScope = keywordScope.Where("d.id IN (SELECT source_id FROM entities WHERE deleted_at IS NULL AND (name ILIKE ? OR aliases::text ILIKE ?))",
			"%"+pq.Entity+"%", "%"+pq.Entity+"%")
		hasCondition = true
	}
	if pq.Type != "" {
		keywordScope = keywordScope.Where("d.type = ?", pq.Type)
		hasCondition = true
	}
	if pq.DateFrom != "" {
		keywordScope = keywordScope.Where("d.created_at >= ?", pq.DateFrom)
		hasCondition = true
	}
	if pq.DateTo != "" {
		keywordScope = keywordScope.Where("d.created_at < ?::date + interval '1 day'", pq.DateTo)
		hasCondition = true
	}
	for _, kw := range pq.Keywords {
		like := "%" + kw + "%"
		keywordScope = keywordScope.Where("(d.content ILIKE ? OR COALESCE(j.result::text,'') ILIKE ?)", like, like)
		hasCondition = true
	}

	hasVec := len(vecUUIDs) > 0

	if !hasCondition && !hasVec {
		return nil, nil
	}

	if hasCondition && hasVec {
		q = q.Where("(d.uuid IN ?) OR (d.id IN (?))", vecUUIDs, keywordScope.Select("d.id"))
	} else if hasVec {
		q = q.Where("d.uuid IN ?", vecUUIDs)
	} else {
		q = keywordScope
	}

	var rows []docRow
	if err := q.Order("d.created_at DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, err
	}

	docIDs := make([]uint, len(rows))
	for i, r := range rows {
		docIDs[i] = r.ID
	}

	entityMap := s.loadEntitiesForDocs(docIDs, tenantID)

	items := make([]SearchItem, 0, len(rows))
	for _, r := range rows {
		item := SearchItem{
			DocumentID:   r.ID,
			DocumentUUID: r.UUID,
			Type:         r.Type,
			SourceURL:    r.SourceURL,
			Content:      r.Content,
			Entities:     entityMap[r.ID],
			CreatedAt:    r.CreatedAt,
		}
		if r.JobResult != nil {
			var jr map[string]any
			if json.Unmarshal(r.JobResult, &jr) == nil {
				if c, ok := jr["content"].(map[string]any); ok {
					item.Summary, _ = c["summary"].(string)
				}
			}
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *SearchService) loadEntitiesForDocs(docIDs []uint, tenantID uint) map[uint][]EntityBrief {
	if len(docIDs) == 0 {
		return nil
	}
	var entities []model.Entity
	q := s.db.Where("source_id IN ? AND deleted_at IS NULL", docIDs)
	if tenantID > 0 {
		q = q.Where("tenant_id = ?", tenantID)
	}
	q.Find(&entities)

	m := make(map[uint][]EntityBrief)
	for _, e := range entities {
		m[e.SourceID] = append(m[e.SourceID], EntityBrief{
			ID: e.ID, Type: e.Type, Name: e.Name,
		})
	}
	return m
}

// --- 阶段 C：智能精排 ---

const rerankSystemPrompt = `你是一个搜索结果精排器。根据用户查询从候选结果中选出最相关的条目。

用户查询：%s
用户意图：%s

候选结果：
%s

请选出最符合用户需求的结果（最多 %d 条），按相关度从高到低排序。
输出 JSON 数组：
[{"index": 0, "score": 9.5, "reason": "推荐理由"}]

index 是候选结果的序号（从 0 开始），score 是 1-10 的相关度评分。
只输出 JSON 数组，不要其他内容。`

func (s *SearchService) rerank(ctx context.Context, query string, pq ParsedQuery, items []SearchItem, limit int) ([]SearchItem, string) {
	topN := limit
	if topN > 10 {
		topN = 10
	}

	var sb strings.Builder
	for i, item := range items {
		entityNames := make([]string, len(item.Entities))
		for j, e := range item.Entities {
			entityNames[j] = e.Name
		}
		fmt.Fprintf(&sb, "[%d] 类型:%s 实体:%s 摘要:%s 时间:%s\n",
			i, item.Type, strings.Join(entityNames, ","),
			truncate(item.Summary, 100), item.CreatedAt.Format("2006-01-02"))
	}

	prompt := fmt.Sprintf(rerankSystemPrompt, query, pq.Intent, sb.String(), topN)
	resp, err := s.gw.Chat(ctx, llm.Request{
		Messages:    []llm.Message{{Role: "user", Content: prompt}},
		Temperature: 0,
	})
	if err != nil {
		logger.Warn("search rerank failed", zap.Error(err))
		return nil, ""
	}

	raw := resp.Content
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start < 0 || end <= start {
		return nil, ""
	}

	var rankings []struct {
		Index  int     `json:"index"`
		Score  float64 `json:"score"`
		Reason string  `json:"reason"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &rankings); err != nil {
		logger.Warn("search rerank JSON parse failed", zap.Error(err))
		return nil, ""
	}

	ranked := make([]SearchItem, 0, len(rankings))
	for _, r := range rankings {
		if r.Index >= 0 && r.Index < len(items) {
			item := items[r.Index]
			item.Score = r.Score
			item.Reason = r.Reason
			ranked = append(ranked, item)
		}
	}
	return ranked, pq.Intent
}

func truncate(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen]) + "..."
}
