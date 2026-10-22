package domain

import "context"

// Repository es el puerto de salida hacia la persistencia.
type Repository interface {
	Create(ctx context.Context, in ConfirmationInput) (Confirmation, error)
	// Update reemplaza nombre y asistentes. Devuelve ErrConfirmationNotFound
	// si la confirmación no existe (por ejemplo, si la eliminó un admin).
	Update(ctx context.Context, id string, in ConfirmationInput) (Confirmation, error)
	List(ctx context.Context) ([]Confirmation, error)
	Delete(ctx context.Context, id string) error
}
