package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/mirainya/nexus/console"
	"github.com/mirainya/nexus/internal/api/handler"
	"github.com/mirainya/nexus/internal/api/middleware"
)

func SetupRouter(asynqClient *asynq.Client) *gin.Engine {
	rl := middleware.NewRateLimiter(100, 200, 10, 20)
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.CORS(), rl.Middleware(), middleware.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	parseH := handler.NewParseHandler()
	jobH := handler.NewJobHandler(asynqClient)
	pipelineH := handler.NewPipelineHandler()
	promptH := handler.NewPromptHandler()
	reviewH := handler.NewReviewHandler()
	entityH := handler.NewEntityHandler()
	authH := handler.NewAuthHandler()
	llmProviderH := handler.NewLLMProviderHandler()
	uploadH := handler.NewUploadHandler()
	sseH := handler.NewSSEHandler()
	searchH := handler.NewSearchHandler()

	// Auth (no middleware)
	r.POST("/api/admin/auth/login", authH.Login)

	// 对外 API（API Key 认证）
	v1 := r.Group("/api/v1", middleware.APIKeyAuth())
	{
		v1.POST("/parse", parseH.Parse)
		v1.POST("/jobs", jobH.Submit)
		v1.GET("/jobs/:id", jobH.GetStatus)
		v1.GET("/jobs/:id/events", sseH.Stream)
		v1.POST("/search", searchH.Search)
	}

	// 管理后台 API（JWT 认证）
	admin := r.Group("/api/admin", middleware.JWTAuth())
	{
		admin.GET("/pipelines", pipelineH.List)
		admin.POST("/pipelines", pipelineH.Create)
		admin.GET("/pipelines/:id", pipelineH.Get)
		admin.PUT("/pipelines/:id", pipelineH.Update)
		admin.DELETE("/pipelines/:id", pipelineH.Delete)

		admin.POST("/pipelines/:id/steps", pipelineH.CreateStep)
		admin.PUT("/pipelines/:id/steps/:step_id", pipelineH.UpdateStep)
		admin.DELETE("/pipelines/:id/steps/:step_id", pipelineH.DeleteStep)
		admin.PUT("/pipelines/:id/steps/reorder", pipelineH.ReorderSteps)

		admin.GET("/prompts", promptH.List)
		admin.POST("/prompts", promptH.Create)
		admin.GET("/prompts/:id", promptH.Get)
		admin.PUT("/prompts/:id", promptH.Update)
		admin.DELETE("/prompts/:id", promptH.Delete)

		admin.GET("/reviews", reviewH.List)
		admin.PUT("/reviews/:id/approve", reviewH.Approve)
		admin.PUT("/reviews/:id/reject", reviewH.Reject)
		admin.PUT("/reviews/:id/modify", reviewH.Modify)

		admin.GET("/entities", entityH.List)
		admin.GET("/entities/:id", entityH.Get)
		admin.GET("/entities/:id/relations", entityH.GetRelations)

		admin.GET("/jobs", jobH.List)
		admin.GET("/jobs/recommend", jobH.Recommend)
		admin.POST("/jobs", jobH.Submit)
		admin.GET("/jobs/:id", jobH.GetStatus)
		admin.GET("/jobs/:id/events", sseH.Stream)
		admin.POST("/jobs/:id/retry", jobH.Retry)

		admin.GET("/llm-providers", llmProviderH.List)
		admin.POST("/llm-providers", llmProviderH.Create)
		admin.PUT("/llm-providers/:id", llmProviderH.Update)
		admin.DELETE("/llm-providers/:id", llmProviderH.Delete)
		admin.PUT("/llm-providers/:id/default", llmProviderH.SetDefault)
		admin.GET("/llm-providers/:name/models", llmProviderH.ListModels)

		admin.POST("/upload", uploadH.Upload)

		admin.POST("/search", searchH.Search)
	}

	console.RegisterRoutes(r)

	return r
}
