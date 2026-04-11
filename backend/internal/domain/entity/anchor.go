package entity

import "time"

type Anchor struct {
	ID           string
	BlockID      string
	SourceNodeID string
	StartOffset  int
	EndOffset    int
	QuotedText   string
	ChildNodeID  *string
	CreatedAt    time.Time
}
