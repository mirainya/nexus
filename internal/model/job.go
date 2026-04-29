package model

import "gorm.io/datatypes"

type Job struct {
	BaseModel
	UUID        string         `gorm:"type:varchar(36);uniqueIndex" json:"uuid"`
	DocumentID  uint           `gorm:"index" json:"document_id"`
	PipelineID  uint           `gorm:"index" json:"pipeline_id"`
	Status      string         `gorm:"type:varchar(20);index;default:pending" json:"status"`
	ContentHash string         `gorm:"type:varchar(64);index" json:"content_hash,omitempty"`
	Result      datatypes.JSON `gorm:"type:jsonb" json:"result"`
	CallbackURL string         `gorm:"type:varchar(1024)" json:"callback_url"`
	CurrentStep int            `gorm:"default:0" json:"current_step"`
	TotalSteps  int            `gorm:"default:0" json:"total_steps"`
	Error       string         `gorm:"type:text" json:"error,omitempty"`
	Document    Document       `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
	Pipeline    Pipeline       `gorm:"foreignKey:PipelineID" json:"pipeline,omitempty"`
	StepLogs    []JobStepLog   `gorm:"foreignKey:JobID" json:"step_logs,omitempty"`
}
