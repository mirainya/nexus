package model

type APIUsage struct {
	BaseModel
	APIKeyID uint   `gorm:"uniqueIndex:idx_apikey_date" json:"api_key_id"`
	Date     string `gorm:"type:varchar(10);uniqueIndex:idx_apikey_date" json:"date"`
	Requests int    `gorm:"default:0" json:"requests"`
	Tokens   int64  `gorm:"default:0" json:"tokens"`
}
