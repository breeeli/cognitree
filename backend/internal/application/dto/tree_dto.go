package dto

import "time"

type CreateTreeRequest struct {
	Title    string `json:"title" binding:"required"`
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

type TreeDetailResponse struct {
	Tree  TreeResponse   `json:"tree"`
	Nodes []NodeResponse `json:"nodes"`
}
