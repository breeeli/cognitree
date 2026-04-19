package dto

type ChatRequest struct {
	Question string `json:"question" binding:"required"`
}

type ChatResponse struct {
	QAPair QAPairResponse `json:"qa_pair"`
}

type ChatStreamEvent struct {
	Type    string          `json:"type"`
	QAPair  *QAPairResponse `json:"qa_pair,omitempty"`
	Delta   string          `json:"delta,omitempty"`
	Message string          `json:"message,omitempty"`
}
