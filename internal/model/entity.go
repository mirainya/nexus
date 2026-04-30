package model

import "gorm.io/datatypes"

type Entity struct {
	BaseModel
	UUID       string         `gorm:"type:varchar(36);uniqueIndex" json:"uuid"`
	Type       string         `gorm:"type:varchar(50);index" json:"type"`
	Name       string         `gorm:"type:varchar(255);index" json:"name"`
	Aliases    datatypes.JSON `gorm:"type:jsonb" json:"aliases"`
	Attributes datatypes.JSON `gorm:"type:jsonb" json:"attributes"`
	Confidence float64        `gorm:"type:decimal(5,4)" json:"confidence"`
	Confirmed  bool           `gorm:"default:false" json:"confirmed"`
	SourceID   uint           `gorm:"index" json:"source_id"`
	Evidence   datatypes.JSON `gorm:"type:jsonb" json:"evidence"`
	TenantID   uint           `gorm:"not null;index" json:"tenant_id"`
}
