package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/minicloud/minicloud/apps/api/internal/config"
	"github.com/minicloud/minicloud/apps/api/internal/database"
	"github.com/minicloud/minicloud/apps/api/internal/logger"
	"github.com/minicloud/minicloud/apps/api/internal/observability"
	"github.com/minicloud/minicloud/apps/api/internal/router"
	"github.com/minicloud/minicloud/apps/api/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := database.Migrate(ctx, db, "apps/api/migrations"); err != nil {
		log.Fatal(err)
	}
	shutdownTelemetry, err := observability.Setup(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = shutdownTelemetry(context.Background()) }()
	logger := logger.New()
	if err := server.Run(ctx, cfg, router.New(cfg, logger, db), logger); err != nil {
		log.Fatal(err)
	}
}
