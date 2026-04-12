package model

import "time"

type Summary struct {
	ID           string `gorm:"type:uuid;primaryKey"`
	TreeID       string `gorm:"type:uuid;not null;index"`
	TargetNodeID string `gorm:"type:uuid;not null;uniqueIndex:uniq_summary_scope_target_version,priority:2;index:idx_summary_scope_target,priority:2"`
	Scope        string `gorm:"type:varchar(20);not null;uniqueIndex:uniq_summary_scope_target_version,priority:1;index:idx_summary_scope_target,priority:1"`
	Version      int64  `gorm:"not null;default:1"`
	AttemptCount int    `gorm:"not null;default:0"`
	Status       string `gorm:"type:varchar(20);not null;default:'pending'"`
	Content      string `gorm:"type:text;not null"`
	ErrorMessage string `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
