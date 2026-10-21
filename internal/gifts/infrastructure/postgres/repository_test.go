package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yosan-aly/regalos-lian-api/internal/gifts/domain"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/database"
)

// Test de integración: requiere una base de datos real. Usa su propio
// schema (TEST_DB_SCHEMAS, por defecto babyshower_test) para no tocar
// los datos de desarrollo.
//
//	TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5433/babyshower?sslmode=disable go test ./...
func newTestPool(t *testing.T) *pgxpool.Pool {
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

// createTestGift crea un regalo temporal y lo borra al terminar el test.
func createTestGift(t *testing.T, pool *pgxpool.Pool, required int) string {
	t.Helper()
	ctx := context.Background()
	id := fmt.Sprintf("test-%d", time.Now().UnixNano())

	if _, err := pool.Exec(ctx, `
		INSERT INTO gifts (id, category_id, name, description, required, sort_order)
		VALUES ($1, 'ropita', 'Regalo de prueba', 'Creado por un test', $2, 999)`, id, required); err != nil {
		t.Fatalf("insert test gift: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM reservations WHERE gift_id = $1`, id)
		_, _ = pool.Exec(ctx, `DELETE FROM gifts WHERE id = $1`, id)
	})
	return id
}

func TestReserve_ConcurrentGuestsNeverOverbook(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)
	giftID := createTestGift(t, pool, 3)

	const guests = 20
	var wg sync.WaitGroup
	var mu sync.Mutex
	ok, soldOut := 0, 0

	for range guests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := repo.Reserve(context.Background(), domain.NewReservation{
				GiftID: giftID, Quantity: 1,
			})
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				ok++
			case errors.Is(err, domain.ErrNotEnoughAvailable):
				soldOut++
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if ok != 3 || soldOut != guests-3 {
		t.Fatalf("ok=%d soldOut=%d, want ok=3 soldOut=%d", ok, soldOut, guests-3)
	}
}

func TestCancelReservation_ReleasesUnits(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)
	ctx := context.Background()
	giftID := createTestGift(t, pool, 2)

	res, gift, err := repo.Reserve(ctx, domain.NewReservation{GiftID: giftID, Quantity: 2})
	if err != nil {
		t.Fatalf("reserve: %v", err)
	}
	if gift.Available() != 0 {
		t.Fatalf("available after reserve = %d, want 0", gift.Available())
	}

	if err := repo.CancelReservation(ctx, res.ID); err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if err := repo.CancelReservation(ctx, res.ID); !errors.Is(err, domain.ErrReservationNotFound) {
		t.Fatalf("second cancel error = %v, want ErrReservationNotFound", err)
	}

	var selected int
	if err := pool.QueryRow(ctx, `SELECT selected FROM gifts WHERE id = $1`, giftID).Scan(&selected); err != nil {
		t.Fatalf("read gift: %v", err)
	}
	if selected != 0 {
		t.Fatalf("selected after cancel = %d, want 0", selected)
	}
}

func TestReserve_UnknownGift(t *testing.T) {
	pool := newTestPool(t)
	repo := NewRepository(pool)

	_, _, err := repo.Reserve(context.Background(), domain.NewReservation{GiftID: "no-existe", Quantity: 1})
	if !errors.Is(err, domain.ErrGiftNotFound) {
		t.Fatalf("error = %v, want ErrGiftNotFound", err)
	}
}
