package model

import "gorm.io/datatypes"

type Relation struct {
	BaseModel
	UUID         string         `gorm:"type:varchar(36);uniqueIndex" json:"uuid"`
	FromEntityID uint           `gorm:"index" json:"from_entity_id"`
	ToEntityID   uint           `gorm:"index" json:"to_entity_id"`
	Type         string         `gorm:"type:varchar(100);index" json:"type"`
	Metadata     datatypes.JSON `gorm:"type:jsonb" json:"metadata"`
	Confidence   float64        `gorm:"type:decimal(5,4)" json:"confidence"`
	Confirmed    bool           `gorm:"default:false" json:"confirmed"`
	SourceID     uint           `gorm:"index" json:"source_id"`
	FromEntity   Entity         `gorm:"foreignKey:FromEntityID" json:"from_entity,omitempty"`
	ToEntity     Entity         `gorm:"foreignKey:ToEntityID" json:"to_entity,omitempty"`
}
