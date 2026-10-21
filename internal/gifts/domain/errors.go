package domain

// Error es un error de negocio con un código estable para el front
// y un mensaje que se puede mostrar directamente al invitado.
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

var (
	ErrGiftNotFound        = &Error{Code: "GIFT_NOT_FOUND", Message: "El regalo no existe."}
	ErrNotEnoughAvailable  = &Error{Code: "NOT_ENOUGH_AVAILABLE", Message: "Ya no quedan suficientes unidades de este regalo. Actualiza la página para ver la disponibilidad."}
	ErrInvalidQuantity     = &Error{Code: "INVALID_QUANTITY", Message: "La cantidad debe estar entre 1 y 20."}
	ErrReservationNotFound = &Error{Code: "RESERVATION_NOT_FOUND", Message: "La reserva no existe o ya fue cancelada."}
)
