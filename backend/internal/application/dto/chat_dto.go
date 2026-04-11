package dto

type ChatRequest struct {
	Question string `json:"question" binding:"required"`
}

type ChatResponse struct {
	QAPair QAPairResponse `json:"qa_pair"`
}
