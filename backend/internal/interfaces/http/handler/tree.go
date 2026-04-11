package handler

import (
	"net/http"

	"github.com/cognitree/backend/internal/application/dto"
	"github.com/cognitree/backend/internal/application/service"
	"github.com/gin-gonic/gin"
)

type TreeHandler struct {
	treeSvc *service.TreeService
}

func NewTreeHandler(treeSvc *service.TreeService) *TreeHandler {
	return &TreeHandler{treeSvc: treeSvc}
}

func (h *TreeHandler) Create(c *gin.Context) {
	var req dto.CreateTreeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.treeSvc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *TreeHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.treeSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *TreeHandler) List(c *gin.Context) {
	resp, err := h.treeSvc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *TreeHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.treeSvc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
