package entity

import "time"

type SummaryScope string

const (
	SummaryScopeNode    SummaryScope = "node"
	SummaryScopePath    SummaryScope = "path"
	SummaryScopeSubtree SummaryScope = "subtree"
)

type SummaryStatus string

const (
	SummaryStatusPending SummaryStatus = "pending"
	SummaryStatusReady   SummaryStatus = "ready"
	SummaryStatusFailed  SummaryStatus = "failed"
)

type Summary struct {
	ID           string
	TreeID       string
	TargetNodeID string
	Scope        SummaryScope
	Version      int64
	AttemptCount int
	Status       SummaryStatus
	Content      string
	ErrorMessage string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
