package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cognitree/backend/internal/application/dto"
	"github.com/cognitree/backend/internal/application/service"
	"github.com/gin-gonic/gin"
)

type TreeStreamHandler struct {
	treeStreamSvc *service.TreeStreamService
}

func NewTreeStreamHandler(treeStreamSvc *service.TreeStreamService) *TreeStreamHandler {
	return &TreeStreamHandler{treeStreamSvc: treeStreamSvc}
}

func (h *TreeStreamHandler) Create(c *gin.Context) {
	var req dto.CreateTreeStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	writeEvent := func(event dto.TreeStreamEvent) error {
		payload, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("marshal sse event: %w", err)
		}

		if _, err := fmt.Fprintf(c.Writer, "event: %s\n", event.Type); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", payload); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	if err := h.treeStreamSvc.StreamFirstQuestion(c.Request.Context(), req, writeEvent); err != nil {
		return
	}
}
