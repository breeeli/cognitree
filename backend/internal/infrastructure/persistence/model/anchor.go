package model

import "time"

type Anchor struct {
	ID           string  `gorm:"type:uuid;primaryKey"`
	BlockID      string  `gorm:"type:uuid;not null;index"`
	SourceNodeID string  `gorm:"type:uuid;not null;index"`
	StartOffset  int     `gorm:"not null"`
	EndOffset    int     `gorm:"not null"`
	QuotedText   string  `gorm:"type:text;not null"`
	ChildNodeID  *string `gorm:"type:uuid"`
	CreatedAt    time.Time
}
