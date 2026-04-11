package model

import "time"

type Node struct {
	ID           string  `gorm:"type:uuid;primaryKey"`
	TreeID       string  `gorm:"type:uuid;not null;index"`
	ParentNodeID *string `gorm:"type:uuid;index"`
	AnchorID     *string `gorm:"type:uuid"`
	Question     string  `gorm:"type:text;not null"`
	Status       string  `gorm:"type:varchar(20);not null;default:'draft'"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
