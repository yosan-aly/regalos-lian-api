package domain

import "context"

// Repository es el puerto de salida hacia la persistencia.
type Repository interface {
	ListCatalog(ctx context.Context) ([]Category, error)
	// Reserve descuenta unidades de forma atómica. Devuelve ErrGiftNotFound
	// o ErrNotEnoughAvailable si no se puede reservar.
	Reserve(ctx context.Context, r NewReservation) (Reservation, Gift, error)
	ListReservations(ctx context.Context, includeCancelled bool) ([]Reservation, error)
	// CancelReservation marca la reserva como cancelada y devuelve las unidades.
	CancelReservation(ctx context.Context, id string) error
}
