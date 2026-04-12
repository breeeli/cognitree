package main

import (
	"context"
	"fmt"
	"os"

	appservice "github.com/cognitree/backend/internal/application/service"
	"github.com/cognitree/backend/internal/infrastructure/ai"
	"github.com/cognitree/backend/internal/infrastructure/config"
	"github.com/cognitree/backend/internal/infrastructure/persistence"
	"github.com/cognitree/backend/internal/infrastructure/persistence/model"
	persistrepo "github.com/cognitree/backend/internal/infrastructure/persistence/repository"
	"github.com/cognitree/backend/internal/interfaces/http/router"
	"github.com/cognitree/backend/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	logger.Init(cfg.Log.Level)
	defer logger.L.Sync()

	logger.L.Info("Connecting to database...")
	db, err := persistence.NewDatabase(cfg.Database)
	if err != nil {
		logger.L.Fatalw("Failed to connect database", "error", err)
	}
	logger.L.Info("Database connected")

	if err := persistence.AutoMigrate(db,
		&model.Tree{},
		&model.Node{},
		&model.QAPair{},
		&model.Block{},
		&model.Anchor{},
		&model.Summary{},
	); err != nil {
		logger.L.Fatalw("Failed to run migrations", "error", err)
	}

	treeRepo := persistrepo.NewTreeRepository(db)
	nodeRepo := persistrepo.NewNodeRepository(db)
	qaPairRepo := persistrepo.NewQAPairRepository(db)
	blockRepo := persistrepo.NewBlockRepository(db)
	anchorRepo := persistrepo.NewAnchorRepository(db)
	summaryRepo := persistrepo.NewSummaryRepository(db)

	treeSvc := appservice.NewTreeService(treeRepo, nodeRepo)
	nodeSvc := appservice.NewNodeService(nodeRepo, qaPairRepo, blockRepo, anchorRepo)

	aiClient := ai.NewOpenAIClient(cfg.AI)
	contextBuilder := ai.NewContextBuilder(treeRepo, nodeRepo, qaPairRepo, blockRepo, anchorRepo, summaryRepo)
	summarySvc := appservice.NewSummaryService(treeRepo, nodeRepo, qaPairRepo, blockRepo, summaryRepo, aiClient)
	summarySvc.Start(context.Background())
	chatSvc := appservice.NewChatService(nodeRepo, qaPairRepo, blockRepo, contextBuilder, aiClient, summarySvc)

	r := router.Setup(router.Deps{
		DB:      db,
		TreeSvc: treeSvc,
		NodeSvc: nodeSvc,
		ChatSvc: chatSvc,
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.L.Infow("Starting server", "addr", addr)
	if err := r.Run(addr); err != nil {
		logger.L.Fatalw("Server failed", "error", err)
	}
}
