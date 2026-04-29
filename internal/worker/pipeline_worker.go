package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/mirainya/nexus/internal/service"
	"github.com/mirainya/nexus/pkg/logger"
	"go.uber.org/zap"
)

const TypePipelineExecute = "pipeline:execute"

type PipelinePayload struct {
	JobID uint `json:"job_id"`
}

func NewPipelineTask(jobID uint) (*asynq.Task, error) {
	payload, _ := json.Marshal(PipelinePayload{JobID: jobID})
	return asynq.NewTask(TypePipelineExecute, payload), nil
}

func HandlePipelineExecute(jobSvc *service.JobService) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p PipelinePayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("unmarshal payload: %w", err)
		}

		logger.Info("executing pipeline job", zap.Uint("job_id", p.JobID))
		return jobSvc.Execute(ctx, p.JobID)
	}
}
