package model

import "time"

type JobStepLog struct {
	BaseModel
	JobID         uint       `gorm:"index" json:"job_id"`
	StepOrder     int        `json:"step_order"`
	ProcessorType string     `gorm:"type:varchar(50)" json:"processor_type"`
	Status        string     `gorm:"type:varchar(20);default:pending" json:"status"`
	StartedAt     *time.Time `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
	Error         string     `gorm:"type:text" json:"error,omitempty"`
	Tokens        int        `gorm:"default:0" json:"tokens"`
	Cost          float64    `gorm:"default:0" json:"cost"`
}
