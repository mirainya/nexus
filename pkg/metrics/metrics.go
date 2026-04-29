package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nexus_http_requests_total",
		Help: "Total HTTP requests",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "nexus_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10, 30, 60, 120},
	}, []string{"method", "path"})

	LLMRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nexus_llm_requests_total",
		Help: "Total LLM requests",
	}, []string{"provider", "model", "status"})

	LLMRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "nexus_llm_request_duration_seconds",
		Help:    "LLM request duration in seconds",
		Buckets: []float64{0.5, 1, 2, 5, 10, 30, 60, 120, 300},
	}, []string{"provider", "model"})

	LLMTokensTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nexus_llm_tokens_total",
		Help: "Total LLM tokens consumed",
	}, []string{"provider", "type"})

	LLMConcurrent = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nexus_llm_concurrent_requests",
		Help: "Current concurrent LLM requests",
	}, []string{"provider"})

	JobsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nexus_jobs_total",
		Help: "Total jobs by status",
	}, []string{"status"})

	JobsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "nexus_jobs_active",
		Help: "Currently active jobs",
	})

	CacheHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nexus_cache_hits_total",
		Help: "Cache hit/miss counts",
	}, []string{"type"})
)
