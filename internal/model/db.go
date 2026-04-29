package model

import "gorm.io/gorm"

var db *gorm.DB

func SetDB(d *gorm.DB) { db = d }
func DB() *gorm.DB     { return db }

func AutoMigrate() error {
	return db.AutoMigrate(
		&User{},
		&APIKey{},
		&PromptTemplate{},
		&Pipeline{},
		&PipelineStep{},
		&Document{},
		&Entity{},
		&Relation{},
		&Job{},
		&JobStepLog{},
		&Review{},
		&LLMProvider{},
	)
}

