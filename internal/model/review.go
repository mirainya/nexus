package model

import "gorm.io/datatypes"

type Review struct {
	BaseModel
	EntityID     *uint          `gorm:"index" json:"entity_id"`
	RelationID   *uint          `gorm:"index" json:"relation_id"`
	Status       string         `gorm:"type:varchar(20);index;default:pending" json:"status"`
	OriginalData datatypes.JSON `gorm:"type:jsonb" json:"original_data"`
	ModifiedData datatypes.JSON `gorm:"type:jsonb" json:"modified_data"`
	Reviewer     string         `gorm:"type:varchar(100)" json:"reviewer"`
}
