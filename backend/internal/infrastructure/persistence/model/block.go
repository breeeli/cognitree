package model

import "time"

type Block struct {
	ID        string `gorm:"type:uuid;primaryKey"`
	QAPairID  string `gorm:"type:uuid;not null;index"`
	Type      string `gorm:"type:varchar(20);not null;default:'paragraph'"`
	Content   string `gorm:"type:text;not null"`
	CreatedAt time.Time
}
