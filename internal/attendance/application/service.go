package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/yosan-aly/regalos-lian-api/internal/attendance/domain"
)

// Service contiene los casos de uso de la confirmación de asistencia.
type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// Confirm registra una nueva confirmación de asistencia.
func (s *Service) Confirm(ctx context.Context, name string, attendees int) (domain.Confirmation, error) {
	in, err := domain.ParseConfirmationInput(name, attendees)
	if err != nil {
		return domain.Confirmation{}, err
	}
	return s.repo.Create(ctx, in)
}

// Update cambia nombre y asistentes. Solo quien tiene el id (guardado en su
// navegador al confirmar) puede actualizar su confirmación.
func (s *Service) Update(ctx context.Context, id, name string, attendees int) (domain.Confirmation, error) {
	if _, err := uuid.Parse(id); err != nil {
		return domain.Confirmation{}, domain.ErrConfirmationNotFound
	}
	in, err := domain.ParseConfirmationInput(name, attendees)
	if err != nil {
		return domain.Confirmation{}, err
	}
	return s.repo.Update(ctx, id, in)
}

// Summary lista las confirmaciones con el total de asistentes (uso administrativo).
func (s *Service) Summary(ctx context.Context) (domain.Summary, error) {
	list, err := s.repo.List(ctx)
	if err != nil {
		return domain.Summary{}, err
	}
	total := 0
	for _, c := range list {
		total += c.Attendees
	}
	return domain.Summary{Confirmations: list, TotalAttendees: total}, nil
}

// Delete elimina una confirmación, por ejemplo un duplicado (uso administrativo).
func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return domain.ErrConfirmationNotFound
	}
	return s.repo.Delete(ctx, id)
}
