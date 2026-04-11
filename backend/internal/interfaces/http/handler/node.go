package handler

import (
	"net/http"

	"github.com/cognitree/backend/internal/application/dto"
	"github.com/cognitree/backend/internal/application/service"
	"github.com/gin-gonic/gin"
)

type NodeHandler struct {
	nodeSvc *service.NodeService
}

func NewNodeHandler(nodeSvc *service.NodeService) *NodeHandler {
	return &NodeHandler{nodeSvc: nodeSvc}
}

func (h *NodeHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.nodeSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *NodeHandler) CreateChild(c *gin.Context) {
	parentID := c.Param("id")

	var req dto.CreateChildNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.nodeSvc.CreateChild(c.Request.Context(), parentID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *NodeHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.nodeSvc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *NodeHandler) GetThread(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.nodeSvc.GetThread(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *NodeHandler) CreateAnchor(c *gin.Context) {
	var req dto.CreateAnchorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.nodeSvc.CreateAnchor(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *NodeHandler) GetAnchors(c *gin.Context) {
	id := c.Param("id")

	resp, err := h.nodeSvc.GetAnchors(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
