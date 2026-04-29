package service

import (
	"fmt"
	"sort"
	"time"

	"github.com/mirainya/nexus/internal/model"
	"gorm.io/gorm"
)

type StatsService struct{ db *gorm.DB }

func NewStatsService(db *gorm.DB) *StatsService { return &StatsService{db: db} }

type DashboardStats struct {
	Jobs       JobStats        `json:"jobs"`
	LLM        LLMStats        `json:"llm"`
	Entities   EntityStats     `json:"entities"`
	DailyTrend []DailyTrendItem `json:"daily_trend"`
}

type JobStats struct {
	Total     int64 `json:"total"`
	Completed int64 `json:"completed"`
	Failed    int64 `json:"failed"`
	Running   int64 `json:"running"`
	Pending   int64 `json:"pending"`
}

type LLMStats struct {
	TotalTokens int64   `json:"total_tokens"`
	TotalCost   float64 `json:"total_cost"`
}

type EntityStats struct {
	Total        int64              `json:"total"`
	Distribution []EntityTypeCount  `json:"distribution"`
}

type EntityTypeCount struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
}

type DailyTrendItem struct {
	Date      string `json:"date"`
	Total     int64  `json:"total"`
	Completed int64  `json:"completed"`
	Failed    int64  `json:"failed"`
}

func (s *StatsService) GetDashboardStats() (*DashboardStats, error) {
	db := s.db
	stats := &DashboardStats{}

	db.Model(&model.Job{}).Count(&stats.Jobs.Total)
	db.Model(&model.Job{}).Where("status = ?", "completed").Count(&stats.Jobs.Completed)
	db.Model(&model.Job{}).Where("status = ?", "failed").Count(&stats.Jobs.Failed)
	db.Model(&model.Job{}).Where("status = ?", "running").Count(&stats.Jobs.Running)
	db.Model(&model.Job{}).Where("status = ?", "pending").Count(&stats.Jobs.Pending)

	var llmResult struct {
		Tokens int64
		Cost   float64
	}
	db.Model(&model.JobStepLog{}).
		Select("COALESCE(SUM(tokens), 0) as tokens, COALESCE(SUM(cost), 0) as cost").
		Scan(&llmResult)
	stats.LLM.TotalTokens = llmResult.Tokens
	stats.LLM.TotalCost = llmResult.Cost

	db.Model(&model.Entity{}).Count(&stats.Entities.Total)
	db.Model(&model.Entity{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&stats.Entities.Distribution)

	since := time.Now().AddDate(0, 0, -6)
	var dailyRows []struct {
		Date      string
		Total     int64
		Completed int64
		Failed    int64
	}
	db.Model(&model.Job{}).
		Select("DATE(created_at) as date, COUNT(*) as total, "+
			"SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed, "+
			"SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed").
		Where("created_at >= ?", since).
		Group("DATE(created_at)").
		Order("date").
		Scan(&dailyRows)

	dateMap := make(map[string]DailyTrendItem)
	for _, r := range dailyRows {
		dateMap[r.Date] = DailyTrendItem{Date: r.Date, Total: r.Total, Completed: r.Completed, Failed: r.Failed}
	}
	for i := 6; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if item, ok := dateMap[d]; ok {
			stats.DailyTrend = append(stats.DailyTrend, item)
		} else {
			stats.DailyTrend = append(stats.DailyTrend, DailyTrendItem{Date: d})
		}
	}

	return stats, nil
}

// --- Pipeline Performance ---

type PipelinePerformance struct {
	PipelineID   uint              `json:"pipeline_id"`
	PipelineName string            `json:"pipeline_name"`
	TotalJobs    int64             `json:"total_jobs"`
	AvgDuration  float64           `json:"avg_duration_ms"`
	P95Duration  float64           `json:"p95_duration_ms"`
	SuccessRate  float64           `json:"success_rate"`
	Steps        []StepPerformance `json:"steps"`
}

type StepPerformance struct {
	ProcessorType string  `json:"processor_type"`
	AvgDuration   float64 `json:"avg_duration_ms"`
	AvgTokens     int     `json:"avg_tokens"`
	AvgCost       float64 `json:"avg_cost"`
	ErrorRate     float64 `json:"error_rate"`
}

func (s *StatsService) GetPipelinePerformance(days int) ([]PipelinePerformance, error) {
	if days <= 0 {
		days = 7
	}
	since := time.Now().AddDate(0, 0, -days)

	var pipelines []model.Pipeline
	s.db.Select("id, name").Find(&pipelines)
	pipelineNames := make(map[uint]string)
	for _, p := range pipelines {
		pipelineNames[p.ID] = p.Name
	}

	type pipelineRow struct {
		PipelineID uint
		Total      int64
		Completed  int64
		AvgMs      float64
	}
	var rows []pipelineRow
	s.db.Model(&model.Job{}).
		Select("pipeline_id, COUNT(*) as total, "+
			"SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed, "+
			"COALESCE(AVG(EXTRACT(EPOCH FROM (updated_at - created_at)) * 1000), 0) as avg_ms").
		Where("created_at >= ?", since).
		Group("pipeline_id").
		Scan(&rows)

	var results []PipelinePerformance
	for _, r := range rows {
		successRate := float64(0)
		if r.Total > 0 {
			successRate = float64(r.Completed) / float64(r.Total) * 100
		}

		var durations []float64
		s.db.Model(&model.Job{}).
			Select("EXTRACT(EPOCH FROM (updated_at - created_at)) * 1000 as dur").
			Where("pipeline_id = ? AND created_at >= ? AND status IN ?", r.PipelineID, since, []string{"completed", "failed"}).
			Pluck("dur", &durations)

		p95 := percentile(durations, 0.95)

		type stepRow struct {
			ProcessorType string
			AvgMs         float64
			AvgTokens     int
			AvgCost       float64
			Total         int64
			Failed        int64
		}
		var stepRows []stepRow
		s.db.Model(&model.JobStepLog{}).
			Select("processor_type, "+
				"COALESCE(AVG(EXTRACT(EPOCH FROM (finished_at - started_at)) * 1000), 0) as avg_ms, "+
				"COALESCE(AVG(tokens), 0) as avg_tokens, "+
				"COALESCE(AVG(cost), 0) as avg_cost, "+
				"COUNT(*) as total, "+
				"SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed").
			Where("job_id IN (SELECT id FROM jobs WHERE pipeline_id = ? AND created_at >= ?)", r.PipelineID, since).
			Group("processor_type").
			Order("processor_type").
			Scan(&stepRows)

		steps := make([]StepPerformance, 0, len(stepRows))
		for _, sr := range stepRows {
			errRate := float64(0)
			if sr.Total > 0 {
				errRate = float64(sr.Failed) / float64(sr.Total) * 100
			}
			steps = append(steps, StepPerformance{
				ProcessorType: sr.ProcessorType,
				AvgDuration:   sr.AvgMs,
				AvgTokens:     sr.AvgTokens,
				AvgCost:       sr.AvgCost,
				ErrorRate:     errRate,
			})
		}

		results = append(results, PipelinePerformance{
			PipelineID:   r.PipelineID,
			PipelineName: pipelineNames[r.PipelineID],
			TotalJobs:    r.Total,
			AvgDuration:  r.AvgMs,
			P95Duration:  p95,
			SuccessRate:  successRate,
			Steps:        steps,
		})
	}

	return results, nil
}

func percentile(data []float64, p float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sort.Float64s(data)
	idx := int(float64(len(data)-1) * p)
	return data[idx]
}

// --- LLM Performance ---

type LLMPerformanceStats struct {
	ByProcessor []ProcessorStats `json:"by_processor"`
	DailyUsage  []DailyLLMUsage  `json:"daily_usage"`
}

type ProcessorStats struct {
	ProcessorType string  `json:"processor_type"`
	TotalCalls    int64   `json:"total_calls"`
	AvgDuration   float64 `json:"avg_duration_ms"`
	TotalTokens   int64   `json:"total_tokens"`
	TotalCost     float64 `json:"total_cost"`
	ErrorRate     float64 `json:"error_rate"`
}

type DailyLLMUsage struct {
	Date   string  `json:"date"`
	Tokens int64   `json:"tokens"`
	Cost   float64 `json:"cost"`
	Calls  int64   `json:"calls"`
}

func (s *StatsService) GetLLMPerformance(days int) (*LLMPerformanceStats, error) {
	if days <= 0 {
		days = 7
	}
	since := time.Now().AddDate(0, 0, -days)

	type procRow struct {
		ProcessorType string
		TotalCalls    int64
		AvgMs         float64
		TotalTokens   int64
		TotalCost     float64
		Failed        int64
	}
	var procRows []procRow
	s.db.Model(&model.JobStepLog{}).
		Select("processor_type, COUNT(*) as total_calls, "+
			"COALESCE(AVG(EXTRACT(EPOCH FROM (finished_at - started_at)) * 1000), 0) as avg_ms, "+
			"COALESCE(SUM(tokens), 0) as total_tokens, "+
			"COALESCE(SUM(cost), 0) as total_cost, "+
			"SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed").
		Where("created_at >= ?", since).
		Group("processor_type").
		Order("total_tokens DESC").
		Scan(&procRows)

	byProc := make([]ProcessorStats, 0, len(procRows))
	for _, r := range procRows {
		errRate := float64(0)
		if r.TotalCalls > 0 {
			errRate = float64(r.Failed) / float64(r.TotalCalls) * 100
		}
		byProc = append(byProc, ProcessorStats{
			ProcessorType: r.ProcessorType,
			TotalCalls:    r.TotalCalls,
			AvgDuration:   r.AvgMs,
			TotalTokens:   r.TotalTokens,
			TotalCost:     r.TotalCost,
			ErrorRate:     errRate,
		})
	}

	type dailyRow struct {
		Date   string
		Tokens int64
		Cost   float64
		Calls  int64
	}
	var dailyRows []dailyRow
	s.db.Model(&model.JobStepLog{}).
		Select("DATE(created_at) as date, COALESCE(SUM(tokens), 0) as tokens, COALESCE(SUM(cost), 0) as cost, COUNT(*) as calls").
		Where("created_at >= ?", since).
		Group("DATE(created_at)").
		Order("date").
		Scan(&dailyRows)

	dateMap := make(map[string]DailyLLMUsage)
	for _, r := range dailyRows {
		dateMap[r.Date] = DailyLLMUsage{Date: r.Date, Tokens: r.Tokens, Cost: r.Cost, Calls: r.Calls}
	}
	daily := make([]DailyLLMUsage, 0, days)
	for i := days - 1; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		if item, ok := dateMap[d]; ok {
			daily = append(daily, item)
		} else {
			daily = append(daily, DailyLLMUsage{Date: d})
		}
	}

	return &LLMPerformanceStats{ByProcessor: byProc, DailyUsage: daily}, nil
}

// --- Error Analysis ---

type ErrorAnalysis struct {
	ErrorTrend     []DailyErrorCount `json:"error_trend"`
	TopErrors      []ErrorGroup      `json:"top_errors"`
	RecentFailures []FailedJobBrief  `json:"recent_failures"`
}

type DailyErrorCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type ErrorGroup struct {
	Error string `json:"error"`
	Count int64  `json:"count"`
}

type FailedJobBrief struct {
	JobID     uint   `json:"job_id"`
	UUID      string `json:"uuid"`
	Error     string `json:"error"`
	Pipeline  string `json:"pipeline"`
	CreatedAt string `json:"created_at"`
}

func (s *StatsService) GetErrorAnalysis(days int) (*ErrorAnalysis, error) {
	if days <= 0 {
		days = 7
	}
	since := time.Now().AddDate(0, 0, -days)

	type trendRow struct {
		Date  string
		Count int64
	}
	var trendRows []trendRow
	s.db.Model(&model.Job{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("status = ? AND created_at >= ?", "failed", since).
		Group("DATE(created_at)").
		Order("date").
		Scan(&trendRows)

	trendMap := make(map[string]int64)
	for _, r := range trendRows {
		trendMap[r.Date] = r.Count
	}
	trend := make([]DailyErrorCount, 0, days)
	for i := days - 1; i >= 0; i-- {
		d := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		trend = append(trend, DailyErrorCount{Date: d, Count: trendMap[d]})
	}

	var topErrors []ErrorGroup
	s.db.Model(&model.Job{}).
		Select("error, COUNT(*) as count").
		Where("status = ? AND created_at >= ? AND error != ''", "failed", since).
		Group("error").
		Order("count DESC").
		Limit(10).
		Scan(&topErrors)

	type failedRow struct {
		ID        uint
		UUID      string
		Error     string
		CreatedAt time.Time
		Pipeline  string
	}
	var failedRows []failedRow
	s.db.Table("jobs j").
		Select("j.id, j.uuid, j.error, j.created_at, COALESCE(p.name, '') as pipeline").
		Joins("LEFT JOIN pipelines p ON p.id = j.pipeline_id").
		Where("j.status = ? AND j.created_at >= ?", "failed", since).
		Order("j.created_at DESC").
		Limit(20).
		Scan(&failedRows)

	recent := make([]FailedJobBrief, 0, len(failedRows))
	for _, r := range failedRows {
		recent = append(recent, FailedJobBrief{
			JobID:     r.ID,
			UUID:      r.UUID,
			Error:     r.Error,
			Pipeline:  r.Pipeline,
			CreatedAt: r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &ErrorAnalysis{
		ErrorTrend:     trend,
		TopErrors:      topErrors,
		RecentFailures: recent,
	}, nil
}

func (s *StatsService) formatDuration(ms float64) string {
	if ms < 1000 {
		return fmt.Sprintf("%.0fms", ms)
	}
	return fmt.Sprintf("%.1fs", ms/1000)
}
