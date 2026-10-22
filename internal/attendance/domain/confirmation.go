package domain

import (
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MinNameLength = 2
	MaxNameLength = 80
	MaxAttendees  = 7
)

// Confirmation registra que un invitado asistirá y con cuántas personas.
type Confirmation struct {
	ID string
	// Name no es único: dos invitados pueden llamarse igual.
	Name string
	// Attendees incluye a quien confirma.
	Attendees int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ConfirmationInput son los datos validados que envía el invitado.
type ConfirmationInput struct {
	Name      string
	Attendees int
}

// Summary agrupa las confirmaciones con sus totales.
type Summary struct {
	Confirmations  []Confirmation
	TotalAttendees int
}

// ParseConfirmationInput valida y normaliza el formulario del invitado.
func ParseConfirmationInput(name string, attendees int) (ConfirmationInput, error) {
	// Colapsa espacios repetidos: "  Sanly   Ortiz " -> "Sanly Ortiz"
	name = strings.Join(strings.Fields(name), " ")
	if n := utf8.RuneCountInString(name); n < MinNameLength || n > MaxNameLength {
		return ConfirmationInput{}, ErrInvalidName
	}
	if attendees < 1 || attendees > MaxAttendees {
		return ConfirmationInput{}, ErrInvalidAttendees
	}
	return ConfirmationInput{Name: name, Attendees: attendees}, nil
}
