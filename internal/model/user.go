package model

type User struct {
	BaseModel
	Username string `gorm:"type:varchar(50);uniqueIndex" json:"username"`
	Password string `gorm:"type:varchar(255)" json:"-"`
	Role     string `gorm:"type:varchar(20);default:admin" json:"role"`
}
