package model

import "gorm.io/datatypes"

type PipelineStep struct {
	BaseModel
	PipelineID       uint            `gorm:"index" json:"pipeline_id"`
	SortOrder        int             `gorm:"default:0" json:"sort_order"`
	ProcessorType    string          `gorm:"type:varchar(50)" json:"processor_type"`
	PromptTemplateID *uint           `json:"prompt_template_id"`
	Config           datatypes.JSON  `gorm:"type:jsonb" json:"config"`
	Condition        string          `gorm:"type:varchar(500)" json:"condition"`
	OnError          string          `gorm:"type:varchar(20);default:stop" json:"on_error"`
	MaxRetry         int             `gorm:"default:0" json:"max_retry"`
	ParallelGroup    int             `gorm:"default:0" json:"parallel_group"`
	PromptTemplate   *PromptTemplate `gorm:"foreignKey:PromptTemplateID" json:"prompt_template,omitempty"`
}
