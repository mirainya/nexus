package model

import "gorm.io/datatypes"

type PromptTemplate struct {
	BaseModel
	Name        string         `gorm:"type:varchar(100)" json:"name"`
	Description string         `gorm:"type:varchar(500)" json:"description"`
	Content     string         `gorm:"type:text" json:"content"`
	Variables   datatypes.JSON `gorm:"type:jsonb" json:"variables"`
	Version     int            `gorm:"default:1" json:"version"`
}
