package application

import (
	"context"
	"testing"

	"github.com/yosan-aly/regalos-lian-api/internal/gifts/domain"
)

// fakeRepo registra las llamadas para verificar que la validación
// ocurre antes de tocar la persistencia.
type fakeRepo struct {
	reserved  []domain.NewReservation
	cancelled []string
}

func (f *fakeRepo) ListCatalog(context.Context) ([]domain.Category, error) { return nil, nil }

func (f *fakeRepo) Reserve(_ context.Context, r domain.NewReservation) (domain.Reservation, domain.Gift, error) {
	f.reserved = append(f.reserved, r)
	return domain.Reservation{GiftID: r.GiftID, Quantity: r.Quantity}, domain.Gift{ID: r.GiftID}, nil
}

func (f *fakeRepo) ListReservations(context.Context, bool) ([]domain.Reservation, error) {
	return nil, nil
}

func (f *fakeRepo) CancelReservation(_ context.Context, id string) error {
	f.cancelled = append(f.cancelled, id)
	return nil
}

func TestReserve_ValidRequestReachesRepository(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	res, _, err := svc.Reserve(context.Background(), " coche ", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.GiftID != "coche" {
		t.Fatalf("GiftID = %q, want %q", res.GiftID, "coche")
	}
	if len(repo.reserved) != 1 {
		t.Fatalf("repo.Reserve called %d times, want 1", len(repo.reserved))
	}
}

func TestReserve_InvalidInputDoesNotReachRepository(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	if _, _, err := svc.Reserve(context.Background(), "coche", 0); err != domain.ErrInvalidQuantity {
		t.Fatalf("error = %v, want ErrInvalidQuantity", err)
	}
	if _, _, err := svc.Reserve(context.Background(), "", 1); err != domain.ErrGiftNotFound {
		t.Fatalf("error = %v, want ErrGiftNotFound", err)
	}
	if len(repo.reserved) != 0 {
		t.Fatalf("repo.Reserve called %d times, want 0", len(repo.reserved))
	}
}

func TestCancelReservation_RejectsInvalidID(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	if err := svc.CancelReservation(context.Background(), "not-a-uuid"); err != domain.ErrReservationNotFound {
		t.Fatalf("error = %v, want ErrReservationNotFound", err)
	}
	if len(repo.cancelled) != 0 {
		t.Fatalf("repo.CancelReservation called %d times, want 0", len(repo.cancelled))
	}

	id := "0b9f1c2e-3d4a-4b5c-8d6e-7f8091a2b3c4"
	if err := svc.CancelReservation(context.Background(), id); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.cancelled) != 1 || repo.cancelled[0] != id {
		t.Fatalf("cancelled = %v, want [%s]", repo.cancelled, id)
	}
}
