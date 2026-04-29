package model

type Pipeline struct {
	BaseModel
	Name        string         `gorm:"type:varchar(100);uniqueIndex" json:"name"`
	Description string         `gorm:"type:varchar(500)" json:"description"`
	Active      bool           `gorm:"default:true" json:"active"`
	Steps       []PipelineStep `gorm:"foreignKey:PipelineID" json:"steps,omitempty"`
}
