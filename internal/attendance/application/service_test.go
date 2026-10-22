package application

import (
	"context"
	"testing"

	"github.com/yosan-aly/regalos-lian-api/internal/attendance/domain"
)

type fakeRepo struct {
	created []domain.ConfirmationInput
	updated []string
	deleted []string
	list    []domain.Confirmation
}

func (f *fakeRepo) Create(_ context.Context, in domain.ConfirmationInput) (domain.Confirmation, error) {
	f.created = append(f.created, in)
	return domain.Confirmation{ID: "id", Name: in.Name, Attendees: in.Attendees}, nil
}

func (f *fakeRepo) Update(_ context.Context, id string, in domain.ConfirmationInput) (domain.Confirmation, error) {
	f.updated = append(f.updated, id)
	return domain.Confirmation{ID: id, Name: in.Name, Attendees: in.Attendees}, nil
}

func (f *fakeRepo) List(context.Context) ([]domain.Confirmation, error) { return f.list, nil }

func (f *fakeRepo) Delete(_ context.Context, id string) error {
	f.deleted = append(f.deleted, id)
	return nil
}

const validID = "0b9f1c2e-3d4a-4b5c-8d6e-7f8091a2b3c4"

func TestConfirm_NormalizesAndValidates(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	got, err := svc.Confirm(context.Background(), "  Ana   María ", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Sanly Ortiz" || got.Attendees != 3 {
		t.Fatalf("got %+v", got)
	}

	if _, err := svc.Confirm(context.Background(), "A", 1); err != domain.ErrInvalidName {
		t.Fatalf("error = %v, want ErrInvalidName", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("repo.Create called %d times, want 1", len(repo.created))
	}
}

func TestUpdate_RejectsInvalidIDAndInput(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	if _, err := svc.Update(context.Background(), "not-a-uuid", "Ana", 2); err != domain.ErrConfirmationNotFound {
		t.Fatalf("error = %v, want ErrConfirmationNotFound", err)
	}
	if _, err := svc.Update(context.Background(), validID, "Ana", 0); err != domain.ErrInvalidAttendees {
		t.Fatalf("error = %v, want ErrInvalidAttendees", err)
	}
	if _, err := svc.Update(context.Background(), validID, "Ana", 2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.updated) != 1 {
		t.Fatalf("repo.Update called %d times, want 1", len(repo.updated))
	}
}

func TestSummary_SumsAttendees(t *testing.T) {
	repo := &fakeRepo{list: []domain.Confirmation{{Attendees: 3}, {Attendees: 1}, {Attendees: 2}}}
	svc := NewService(repo)

	got, err := svc.Summary(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TotalAttendees != 6 || len(got.Confirmations) != 3 {
		t.Fatalf("got total=%d count=%d, want total=6 count=3", got.TotalAttendees, len(got.Confirmations))
	}
}

func TestDelete_RejectsInvalidID(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(repo)

	if err := svc.Delete(context.Background(), "nope"); err != domain.ErrConfirmationNotFound {
		t.Fatalf("error = %v, want ErrConfirmationNotFound", err)
	}
	if len(repo.deleted) != 0 {
		t.Fatalf("repo.Delete called %d times, want 0", len(repo.deleted))
	}
}
