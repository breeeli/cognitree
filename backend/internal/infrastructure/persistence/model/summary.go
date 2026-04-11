package model

import "time"

type Summary struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	NodeID    string `gorm:"type:uuid;not null;index"`
	Type      string `gorm:"type:varchar(20);not null"`
	Content   string `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
