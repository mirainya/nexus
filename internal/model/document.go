package model

import "gorm.io/datatypes"

type Document struct {
	BaseModel
	UUID      string         `gorm:"type:varchar(36);uniqueIndex" json:"uuid"`
	Type      string         `gorm:"type:varchar(50);index" json:"type"`
	Content   string         `gorm:"type:text" json:"content"`
	SourceURL string         `gorm:"type:varchar(1024)" json:"source_url"`
	Metadata  datatypes.JSON `gorm:"type:jsonb" json:"metadata"`
	Status    string         `gorm:"type:varchar(20);index;default:pending" json:"status"`
	FilePath  string         `gorm:"type:varchar(512)" json:"file_path"`
	TenantID  uint           `gorm:"not null;index" json:"tenant_id"`
}
