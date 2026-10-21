package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/yosan-aly/regalos-lian-api/internal/gifts/application"
	"github.com/yosan-aly/regalos-lian-api/internal/gifts/infrastructure/postgres"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/config"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/database"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/logger"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped", "error", err.Error())
		os.Exit(1)
	}
}

func run() error {
	// Solo en local: .env (secretos, ignorado por git) y luego .dev.env.
	// godotenv no sobrescribe variables ya definidas, así que en Cloud Run
	// (donde estos archivos no existen) mandan las del entorno.
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".dev.env")

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.LogLevel)
	slog.SetDefault(log)

	// Cloud Run envía SIGTERM antes de apagar la instancia.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DB)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool); err != nil {
		return err
	}

	svc := application.NewService(postgres.NewRepository(pool))
	app := server.New(cfg, svc, log)

	errCh := make(chan error, 1)
	go func() { errCh <- app.Listen(":" + cfg.Port) }()
	log.Info("server started", "port", cfg.Port, "adminEnabled", cfg.AdminToken != "")

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	log.Info("shutting down")
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}
