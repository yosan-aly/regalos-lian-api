// Package dbtest ayuda a los tests de integración a usar una base real.
package dbtest

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yosan-aly/regalos-lian-api/internal/platform/database"
)

// NewPool conecta a TEST_DATABASE_URL dentro del schema TEST_DB_SCHEMAS
// (por defecto babyshower_test) y aplica las migraciones. Si la variable
// no está definida, el test se omite.
//
//	TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5433/babyshower?sslmode=disable go test ./...
func NewPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL no está definida; se omite el test de integración")
	}
	schema := os.Getenv("TEST_DB_SCHEMAS")
	if schema == "" {
		schema = "babyshower_test"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := database.Migrate(ctx, pool, schema); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return pool
}
