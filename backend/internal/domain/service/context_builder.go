package service

import "context"

type ContextPayload struct {
	SystemPrompt string
	UserPrompt   string
	Degraded     bool
	Warnings     []string
}

type ContextBuilder interface {
	BuildContext(ctx context.Context, treeID string, currentNodeID string, newQuestion string) (*ContextPayload, error)
}
