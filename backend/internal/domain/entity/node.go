package entity

import "time"

type NodeStatus string

const (
	NodeStatusDraft      NodeStatus = "draft"
	NodeStatusAnswered   NodeStatus = "answered"
	NodeStatusSummarized NodeStatus = "summarized"
)

type Node struct {
	ID           string
	TreeID       string
	ParentNodeID *string
	AnchorID     *string
	Question     string
	Status       NodeStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	QAPairs      []QAPair
}
