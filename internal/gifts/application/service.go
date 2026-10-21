package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/yosan-aly/regalos-lian-api/internal/gifts/domain"
)

// Service contiene los casos de uso de la lista de regalos.
type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// Catalog devuelve las categorías con sus regalos y disponibilidad.
func (s *Service) Catalog(ctx context.Context) ([]domain.Category, error) {
	return s.repo.ListCatalog(ctx)
}

// Reserve valida la solicitud y reserva las unidades.
func (s *Service) Reserve(ctx context.Context, giftID string, quantity int) (domain.Reservation, domain.Gift, error) {
	req, err := domain.ParseNewReservation(giftID, quantity)
	if err != nil {
		return domain.Reservation{}, domain.Gift{}, err
	}
	return s.repo.Reserve(ctx, req)
}

// Reservations lista las reservas (uso administrativo).
func (s *Service) Reservations(ctx context.Context, includeCancelled bool) ([]domain.Reservation, error) {
	return s.repo.ListReservations(ctx, includeCancelled)
}

// CancelReservation libera las unidades de una reserva (uso administrativo).
func (s *Service) CancelReservation(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return domain.ErrReservationNotFound
	}
	return s.repo.CancelReservation(ctx, id)
}
