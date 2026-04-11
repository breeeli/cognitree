package dto

import "time"

type CreateChildNodeRequest struct {
	Question string  `json:"question" binding:"required"`
	AnchorID *string `json:"anchor_id,omitempty"`
}

type NodeResponse struct {
	ID           string           `json:"id"`
	TreeID       string           `json:"tree_id"`
	ParentNodeID *string          `json:"parent_node_id"`
	AnchorID     *string          `json:"anchor_id"`
	Question     string           `json:"question"`
	Status       string           `json:"status"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	QAPairs      []QAPairResponse `json:"qa_pairs,omitempty"`
}

type QAPairResponse struct {
	ID        string          `json:"id"`
	Question  string          `json:"question"`
	Blocks    []BlockResponse `json:"blocks"`
	CreatedAt time.Time       `json:"created_at"`
}

type BlockResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Content string `json:"content"`
}

type ThreadResponse struct {
	Nodes []NodeResponse `json:"nodes"`
}

type CreateAnchorRequest struct {
	BlockID       string `json:"block_id" binding:"required"`
	StartOffset   int    `json:"start_offset" binding:"required"`
	EndOffset     int    `json:"end_offset" binding:"required"`
	QuotedText    string `json:"quoted_text" binding:"required"`
	ChildQuestion string `json:"child_question" binding:"required"`
}

type AnchorResponse struct {
	ID           string  `json:"id"`
	BlockID      string  `json:"block_id"`
	SourceNodeID string  `json:"source_node_id"`
	StartOffset  int     `json:"start_offset"`
	EndOffset    int     `json:"end_offset"`
	QuotedText   string  `json:"quoted_text"`
	ChildNodeID  *string `json:"child_node_id"`
}

type CreateAnchorResponse struct {
	Anchor    AnchorResponse `json:"anchor"`
	ChildNode NodeResponse   `json:"child_node"`
}
