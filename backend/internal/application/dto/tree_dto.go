package dto

import "time"

type CreateTreeRequest struct {
	Title    string `json:"title,omitempty"`
	Question string `json:"question" binding:"required"`
}

type TreeResponse struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateTreeResponse struct {
	Tree     TreeResponse `json:"tree"`
	RootNode NodeResponse `json:"root_node"`
}

type CreateTreeStreamRequest struct {
	TreeID     string `json:"tree_id,omitempty"`
	RootNodeID string `json:"root_node_id,omitempty"`
	Question   string `json:"question" binding:"required"`
}

type TreeStreamEvent struct {
	Type     string          `json:"type"`
	Tree     *TreeResponse   `json:"tree,omitempty"`
	RootNode *NodeResponse   `json:"root_node,omitempty"`
	QAPair   *QAPairResponse `json:"qa_pair,omitempty"`
	Delta    string          `json:"delta,omitempty"`
	Message  string          `json:"message,omitempty"`
}

type TreeDetailResponse struct {
	Tree  TreeResponse   `json:"tree"`
	Nodes []NodeResponse `json:"nodes"`
}
