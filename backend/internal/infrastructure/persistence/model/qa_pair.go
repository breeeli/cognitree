package model

import "time"

type QAPair struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	NodeID    string `gorm:"type:uuid;not null;index"`
	Question  string `gorm:"type:text;not null"`
	CreatedAt time.Time
}
