package domain

// Error es un error de negocio con un código estable para el front
// y un mensaje que se puede mostrar directamente al invitado.
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

var (
	ErrInvalidName          = &Error{Code: "INVALID_NAME", Message: "Escribe tu nombre (entre 2 y 80 caracteres)."}
	ErrInvalidAttendees     = &Error{Code: "INVALID_ATTENDEES", Message: "La cantidad de asistentes debe estar entre 1 y 7."}
	ErrConfirmationNotFound = &Error{Code: "CONFIRMATION_NOT_FOUND", Message: "La confirmación no existe."}
)
