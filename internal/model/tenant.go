package model

type Tenant struct {
	BaseModel
	UUID   string `gorm:"type:varchar(36);uniqueIndex" json:"uuid"`
	Name   string `gorm:"type:varchar(100);uniqueIndex" json:"name"`
	Active bool   `gorm:"default:true" json:"active"`
}
