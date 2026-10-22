package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/yosan-aly/regalos-lian-api/internal/attendance/domain"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/database/dbtest"
)

func TestConfirmation_CreateUpdateListDelete(t *testing.T) {
	pool := dbtest.NewPool(t)
	repo := NewRepository(pool)
	ctx := context.Background()

	created, err := repo.Create(ctx, domain.ConfirmationInput{Name: "Invitado de prueba", Attendees: 2})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), created.ID) })

	updated, err := repo.Update(ctx, created.ID, domain.ConfirmationInput{Name: "Invitado actualizado", Attendees: 4})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Invitado actualizado" || updated.Attendees != 4 || updated.ID != created.ID {
		t.Fatalf("updated = %+v", updated)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	found := false
	for _, c := range list {
		if c.ID == created.ID {
			found = c.Attendees == 4
		}
	}
	if !found {
		t.Fatalf("confirmation %s not listed with updated attendees", created.ID)
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := repo.Delete(ctx, created.ID); !errors.Is(err, domain.ErrConfirmationNotFound) {
		t.Fatalf("second delete error = %v, want ErrConfirmationNotFound", err)
	}
	if _, err := repo.Update(ctx, created.ID, domain.ConfirmationInput{Name: "Ana", Attendees: 1}); !errors.Is(err, domain.ErrConfirmationNotFound) {
		t.Fatalf("update after delete error = %v, want ErrConfirmationNotFound", err)
	}
}

func TestConfirmation_DuplicateNamesAllowed(t *testing.T) {
	pool := dbtest.NewPool(t)
	repo := NewRepository(pool)
	ctx := context.Background()

	for range 2 {
		c, err := repo.Create(ctx, domain.ConfirmationInput{Name: "Nombre repetido", Attendees: 1})
		if err != nil {
			t.Fatalf("create duplicate name: %v", err)
		}
		t.Cleanup(func() { _ = repo.Delete(context.Background(), c.ID) })
	}
}
