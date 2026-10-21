package domain

import (
	"strings"
	"time"
)

const MaxQuantity = 20

// Reservation registra qué regalo se eligió y cuántas unidades.
// Es anónima: no guarda datos del invitado.
type Reservation struct {
	ID          string
	GiftID      string
	GiftName    string
	GiftVariant string
	Quantity    int
	CreatedAt   time.Time
	CancelledAt *time.Time
}

// NewReservation es una solicitud de reserva ya validada.
type NewReservation struct {
	GiftID   string
	Quantity int
}

// ParseNewReservation valida los datos de la solicitud.
func ParseNewReservation(giftID string, quantity int) (NewReservation, error) {
	giftID = strings.TrimSpace(giftID)
	if giftID == "" {
		return NewReservation{}, ErrGiftNotFound
	}
	if quantity < 1 || quantity > MaxQuantity {
		return NewReservation{}, ErrInvalidQuantity
	}
	return NewReservation{GiftID: giftID, Quantity: quantity}, nil
}
