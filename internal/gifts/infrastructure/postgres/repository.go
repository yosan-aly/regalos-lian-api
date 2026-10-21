package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yosan-aly/regalos-lian-api/internal/gifts/domain"
)

// Repository implementa domain.Repository sobre PostgreSQL.
type Repository struct {
	pool *pgxpool.Pool
}

var _ domain.Repository = (*Repository)(nil)

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListCatalog(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.title, c.description,
		       g.id, g.name, COALESCE(g.variant, ''), g.description, g.required, g.selected
		FROM categories c
		JOIN gifts g ON g.category_id = c.id
		ORDER BY c.sort_order, g.sort_order`)
	if err != nil {
		return nil, fmt.Errorf("query catalog: %w", err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var c domain.Category
		var g domain.Gift
		if err := rows.Scan(&c.ID, &c.Title, &c.Description,
			&g.ID, &g.Name, &g.Variant, &g.Description, &g.Required, &g.Selected); err != nil {
			return nil, fmt.Errorf("scan catalog: %w", err)
		}
		g.CategoryID = c.ID

		// Las filas vienen ordenadas por categoría: se agrupan al vuelo.
		if n := len(categories); n == 0 || categories[n-1].ID != c.ID {
			categories = append(categories, c)
		}
		last := &categories[len(categories)-1]
		last.Gifts = append(last.Gifts, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read catalog: %w", err)
	}
	return categories, nil
}

func (r *Repository) Reserve(ctx context.Context, req domain.NewReservation) (domain.Reservation, domain.Gift, error) {
	var res domain.Reservation
	var gift domain.Gift

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		// El UPDATE condicional bloquea la fila: si dos invitados reservan
		// a la vez, el segundo reevalúa la condición con el valor ya actualizado.
		err := tx.QueryRow(ctx, `
			UPDATE gifts
			SET selected = selected + $2, updated_at = now()
			WHERE id = $1 AND selected + $2 <= required
			RETURNING id, category_id, name, COALESCE(variant, ''), description, required, selected`,
			req.GiftID, req.Quantity,
		).Scan(&gift.ID, &gift.CategoryID, &gift.Name, &gift.Variant, &gift.Description, &gift.Required, &gift.Selected)

		if errors.Is(err, pgx.ErrNoRows) {
			var exists bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM gifts WHERE id = $1)`, req.GiftID).Scan(&exists); err != nil {
				return fmt.Errorf("check gift: %w", err)
			}
			if !exists {
				return domain.ErrGiftNotFound
			}
			return domain.ErrNotEnoughAvailable
		}
		if err != nil {
			return fmt.Errorf("update gift: %w", err)
		}

		err = tx.QueryRow(ctx, `
			INSERT INTO reservations (gift_id, quantity)
			VALUES ($1, $2)
			RETURNING id::text, created_at`,
			req.GiftID, req.Quantity,
		).Scan(&res.ID, &res.CreatedAt)
		if err != nil {
			return fmt.Errorf("insert reservation: %w", err)
		}
		return nil
	})
	if err != nil {
		return domain.Reservation{}, domain.Gift{}, err
	}

	res.GiftID = gift.ID
	res.GiftName = gift.Name
	res.GiftVariant = gift.Variant
	res.Quantity = req.Quantity
	return res, gift, nil
}

func (r *Repository) ListReservations(ctx context.Context, includeCancelled bool) ([]domain.Reservation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT r.id::text, r.gift_id, g.name, COALESCE(g.variant, ''),
		       r.quantity, r.created_at, r.cancelled_at
		FROM reservations r
		JOIN gifts g ON g.id = r.gift_id
		WHERE $1 OR r.cancelled_at IS NULL
		ORDER BY r.created_at DESC`, includeCancelled)
	if err != nil {
		return nil, fmt.Errorf("query reservations: %w", err)
	}
	defer rows.Close()

	var list []domain.Reservation
	for rows.Next() {
		var res domain.Reservation
		if err := rows.Scan(&res.ID, &res.GiftID, &res.GiftName, &res.GiftVariant,
			&res.Quantity, &res.CreatedAt, &res.CancelledAt); err != nil {
			return nil, fmt.Errorf("scan reservation: %w", err)
		}
		list = append(list, res)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read reservations: %w", err)
	}
	return list, nil
}

func (r *Repository) CancelReservation(ctx context.Context, id string) error {
	return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		var giftID string
		var quantity int
		err := tx.QueryRow(ctx, `
			UPDATE reservations
			SET cancelled_at = now()
			WHERE id = $1 AND cancelled_at IS NULL
			RETURNING gift_id, quantity`, id,
		).Scan(&giftID, &quantity)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrReservationNotFound
		}
		if err != nil {
			return fmt.Errorf("cancel reservation: %w", err)
		}

		if _, err := tx.Exec(ctx, `
			UPDATE gifts
			SET selected = selected - $2, updated_at = now()
			WHERE id = $1`, giftID, quantity); err != nil {
			return fmt.Errorf("release gift units: %w", err)
		}
		return nil
	})
}
