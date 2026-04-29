package worker

import (
	"github.com/hibiken/asynq"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/config"
)

func NewServer() *asynq.Server {
	cfg := config.C.Redis
	concurrency := config.C.Worker.Concurrency
	if concurrency <= 0 {
		concurrency = 10
	}
	return asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		},
		asynq.Config{Concurrency: concurrency},
	)
}

func NewClient() *asynq.Client {
	cfg := config.C.Redis
	return asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}

func RegisterHandlers(mux *asynq.ServeMux, jobSvc *service.JobService) {
	mux.HandleFunc(TypePipelineExecute, HandlePipelineExecute(jobSvc))
}
