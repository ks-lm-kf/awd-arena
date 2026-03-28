package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/awd-platform/awd-arena/internal/config"
	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/engine"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/internal/server"
	"github.com/awd-platform/awd-arena/internal/service"
	"github.com/awd-platform/awd-arena/pkg/crypto"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

func main() {
	cfgPath := "configs/config.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	logger.Init("info")
	log := logger.Get()

	if err := database.InitDB(cfg.Database.SQLitePath); err != nil {
		log.Error("database init error", "error", err)
		os.Exit(1)
	}
	log.Info("database initialized", "path", cfg.Database.SQLitePath)

	// Auto-migrate all models
	db := database.GetDB()
	db.AutoMigrate(
		&model.User{},
		&model.Team{},
		&model.Game{},
		&model.Challenge{},
		&model.TeamContainer{},
		&model.FlagRecord{},
		&model.FlagSubmission{},
		&model.RoundScore{},
		&model.EventLog{},
		&model.AdminLog{},
		&model.ServiceHealth{},
	)
	log.Info("database migrated")

	seedAdmin()

	// Set up engine callbacks to avoid cyclic dependency
	service.EngineCallbacks.StartGame = func(game *model.Game) error {
		return engine.Manager.StartGame(game)
	}
	service.EngineCallbacks.PauseGame = func(gameID int64) error {
		return engine.Manager.PauseGame(gameID)
	}
	service.EngineCallbacks.ResumeGame = func(game *model.Game) error {
		return engine.Manager.ResumeGame(game)
	}
	service.EngineCallbacks.StopGame = func(gameID int64) error {
		return engine.Manager.StopGame(gameID)
	}
	log.Info("engine callbacks initialized")

	srv := server.New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Start(); err != nil {
			log.Error("server error", "error", err)
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down...")
	engine.Manager.ShutdownAll()
	srv.Shutdown(ctx)
}

func seedAdmin() {
	db := database.GetDB()
	var count int64
	db.Model(&model.User{}).Where("role = ?", "admin").Count(&count)
	if count == 0 {
		hashed, err := crypto.HashPassword("admin123")
		if err != nil {
			logger.Get().Error("failed to hash admin password", "error", err)
			return
		}
		admin := model.User{
			Username: "admin",
			Password: hashed,
			Role:     "admin",
		}
		if err := db.Create(&admin).Error; err != nil {
			logger.Get().Error("failed to create admin", "error", err)
			return
		}
		logger.Get().Info("Default admin account created: admin/admin123")
	}
}
