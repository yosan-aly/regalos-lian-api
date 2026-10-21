package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yosan-aly/regalos-lian-api/internal/platform/config"
)

// Connect abre el pool de conexiones y verifica que la base responda.
func Connect(ctx context.Context, cfg config.DB) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL())
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}
	// Pocas conexiones y sin mínimas: Cloud Run escala a 0 y Neon
	// suspende el cómputo cuando no hay conexiones abiertas.
	poolCfg.MaxConns = 4
	poolCfg.MinConns = 0
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	// Cada conexión trabaja dentro del schema del servicio. Es un parámetro
	// de arranque de la conexión: requiere el host directo de Neon, no el -pooler.
	poolCfg.ConnConfig.RuntimeParams["search_path"] = cfg.Schema

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Neon puede tardar unos segundos en despertar.
	pingCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}
