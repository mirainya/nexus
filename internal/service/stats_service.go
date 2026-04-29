package service

import (
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
