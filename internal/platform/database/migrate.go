package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate crea el schema si no existe y aplica en orden los archivos de
// migrations/ que aún no se han ejecutado en él. Cada archivo corre en su
// propia transacción. Las tablas se crean en el schema porque el pool
// fija search_path (ver Connect).
func Migrate(ctx context.Context, pool *pgxpool.Pool, schema string) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	// Advisory lock por schema: evita que dos instancias de Cloud Run
	// migren el mismo schema al mismo tiempo.
	lockKey := "regalos-lian-api:" + schema
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock(hashtext($1))`, lockKey); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer conn.Exec(context.Background(), `SELECT pg_advisory_unlock(hashtext($1))`, lockKey) //nolint:errcheck

	// Se consulta antes de crear: CREATE SCHEMA exige permiso CREATE sobre la
	// base incluso con IF NOT EXISTS, y un usuario de ejecución puede no tenerlo
	// si el schema ya fue creado por el administrador.
	var exists bool
	if err := conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = $1)`, schema).Scan(&exists); err != nil {
		return fmt.Errorf("check schema: %w", err)
	}
	if !exists {
		if _, err := conn.Exec(ctx, `CREATE SCHEMA `+pgx.Identifier{schema}.Sanitize()); err != nil {
			return fmt.Errorf("create schema %s: %w", schema, err)
		}
		slog.InfoContext(ctx, "schema created", "schema", schema)
	}

	if _, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    text PRIMARY KEY,
			applied_at timestamptz NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	// ReadDir devuelve los archivos ordenados por nombre (0001_, 0002_, ...).
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	for _, entry := range entries {
		version := entry.Name()

		var applied bool
		if err := conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if applied {
			continue
		}

		script, err := migrationsFS.ReadFile("migrations/" + version)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}

		err = pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
			// Sin argumentos, pgx usa el protocolo simple y admite varias sentencias.
			if _, err := tx.Exec(ctx, string(script)); err != nil {
				return err
			}
			_, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version)
			return err
		})
		if err != nil {
			return fmt.Errorf("apply migration %s: %w", version, err)
		}
		slog.InfoContext(ctx, "migration applied", "schema", schema, "version", version)
	}
	return nil
}
