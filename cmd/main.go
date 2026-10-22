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

	attendanceapp "github.com/yosan-aly/regalos-lian-api/internal/attendance/application"
	attendancepg "github.com/yosan-aly/regalos-lian-api/internal/attendance/infrastructure/postgres"
	giftsapp "github.com/yosan-aly/regalos-lian-api/internal/gifts/application"
	giftspg "github.com/yosan-aly/regalos-lian-api/internal/gifts/infrastructure/postgres"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/config"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/database"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/logger"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/server"
)

// @title Regalos de Lian API
// @version 1.0.0
// @description Lista de regalos del baby shower de Lian. Los invitados reservan
// @description regalos de forma anónima y la disponibilidad se actualiza para todos.
// @host localhost:7070
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey AdminToken
// @in header
// @name X-Admin-Token
// @description Token de administración (variable ADMIN_TOKEN)
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

	if err := database.Migrate(ctx, pool, cfg.DB.Schema); err != nil {
		return err
	}

	app := server.New(cfg, server.Services{
		Gifts:      giftsapp.NewService(giftspg.NewRepository(pool)),
		Attendance: attendanceapp.NewService(attendancepg.NewRepository(pool)),
	}, log)

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
