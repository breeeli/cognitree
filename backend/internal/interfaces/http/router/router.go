package router

import (
	"github.com/cognitree/backend/internal/application/service"
	"github.com/cognitree/backend/internal/interfaces/http/handler"
	"github.com/cognitree/backend/internal/interfaces/http/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Deps struct {
	DB      *gorm.DB
	TreeSvc *service.TreeService
	NodeSvc *service.NodeService
	ChatSvc *service.ChatService
}

func Setup(deps Deps) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS())

	healthHandler := handler.NewHealthHandler(deps.DB)
	treeHandler := handler.NewTreeHandler(deps.TreeSvc)
	nodeHandler := handler.NewNodeHandler(deps.NodeSvc)
	chatHandler := handler.NewChatHandler(deps.ChatSvc)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", healthHandler.Check)

		v1.POST("/trees", treeHandler.Create)
		v1.GET("/trees", treeHandler.List)
		v1.GET("/trees/:id", treeHandler.GetByID)
		v1.DELETE("/trees/:id", treeHandler.Delete)

		v1.GET("/nodes/:id", nodeHandler.GetByID)
		v1.POST("/nodes/:id/children", nodeHandler.CreateChild)
		v1.DELETE("/nodes/:id", nodeHandler.Delete)
		v1.GET("/nodes/:id/thread", nodeHandler.GetThread)
		v1.GET("/nodes/:id/anchors", nodeHandler.GetAnchors)

		v1.POST("/anchors", nodeHandler.CreateAnchor)

		v1.POST("/nodes/:id/chat", chatHandler.Chat)
	}

	return r
}
