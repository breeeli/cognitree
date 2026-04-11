package model

import "time"

type Tree struct {
	ID          string `gorm:"type:uuid;primaryKey"`
	Title       string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
