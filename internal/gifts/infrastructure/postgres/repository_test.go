package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yosan-aly/regalos-lian-api/internal/gifts/domain"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/database/dbtest"
)

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
	pool := dbtest.NewPool(t)
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
	pool := dbtest.NewPool(t)
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
	pool := dbtest.NewPool(t)
	repo := NewRepository(pool)

	_, _, err := repo.Reserve(context.Background(), domain.NewReservation{GiftID: "no-existe", Quantity: 1})
	if !errors.Is(err, domain.ErrGiftNotFound) {
		t.Fatalf("error = %v, want ErrGiftNotFound", err)
	}
}
