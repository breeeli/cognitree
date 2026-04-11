package handler

import (
	"net/http"

	"github.com/cognitree/backend/internal/application/dto"
	"github.com/cognitree/backend/internal/application/service"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatSvc *service.ChatService
}

func NewChatHandler(chatSvc *service.ChatService) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc}
}

func (h *ChatHandler) Chat(c *gin.Context) {
	nodeID := c.Param("id")

	var req dto.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.chatSvc.Chat(c.Request.Context(), nodeID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
