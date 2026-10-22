package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yosan-aly/regalos-lian-api/internal/attendance/domain"
)

// Repository implementa domain.Repository sobre PostgreSQL.
type Repository struct {
	pool *pgxpool.Pool
}

var _ domain.Repository = (*Repository)(nil)

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const confirmationColumns = `id::text, name, attendees, created_at, updated_at`

func scanConfirmation(row pgx.Row) (domain.Confirmation, error) {
	var c domain.Confirmation
	err := row.Scan(&c.ID, &c.Name, &c.Attendees, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *Repository) Create(ctx context.Context, in domain.ConfirmationInput) (domain.Confirmation, error) {
	c, err := scanConfirmation(r.pool.QueryRow(ctx, `
		INSERT INTO attendance_confirmations (name, attendees)
		VALUES ($1, $2)
		RETURNING `+confirmationColumns, in.Name, in.Attendees))
	if err != nil {
		return domain.Confirmation{}, fmt.Errorf("insert confirmation: %w", err)
	}
	return c, nil
}

func (r *Repository) Update(ctx context.Context, id string, in domain.ConfirmationInput) (domain.Confirmation, error) {
	c, err := scanConfirmation(r.pool.QueryRow(ctx, `
		UPDATE attendance_confirmations
		SET name = $2, attendees = $3, updated_at = now()
		WHERE id = $1
		RETURNING `+confirmationColumns, id, in.Name, in.Attendees))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Confirmation{}, domain.ErrConfirmationNotFound
	}
	if err != nil {
		return domain.Confirmation{}, fmt.Errorf("update confirmation: %w", err)
	}
	return c, nil
}

func (r *Repository) List(ctx context.Context) ([]domain.Confirmation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+confirmationColumns+`
		FROM attendance_confirmations
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query confirmations: %w", err)
	}
	defer rows.Close()

	var list []domain.Confirmation
	for rows.Next() {
		c, err := scanConfirmation(rows)
		if err != nil {
			return nil, fmt.Errorf("scan confirmation: %w", err)
		}
		list = append(list, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read confirmations: %w", err)
	}
	return list, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM attendance_confirmations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete confirmation: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConfirmationNotFound
	}
	return nil
}
