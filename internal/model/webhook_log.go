package model

type WebhookLog struct {
	BaseModel
	JobID      uint   `gorm:"index" json:"job_id"`
	URL        string `gorm:"type:varchar(1024)" json:"url"`
	Event      string `gorm:"type:varchar(50)" json:"event"`
	Status     string `gorm:"type:varchar(20)" json:"status"`
	StatusCode int    `json:"status_code"`
	Error      string `gorm:"type:text" json:"error,omitempty"`
	Attempts   int    `json:"attempts"`
}
