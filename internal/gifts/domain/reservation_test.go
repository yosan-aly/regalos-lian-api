package domain

import "testing"

func TestParseNewReservation(t *testing.T) {
	tests := []struct {
		name     string
		giftID   string
		quantity int
		want     NewReservation
		wantErr  error
	}{
		{name: "valid", giftID: "coche", quantity: 1, want: NewReservation{GiftID: "coche", Quantity: 1}},
		{name: "trims gift id", giftID: " coche ", quantity: 2, want: NewReservation{GiftID: "coche", Quantity: 2}},
		{name: "empty gift", giftID: "  ", quantity: 1, wantErr: ErrGiftNotFound},
		{name: "zero quantity", giftID: "coche", quantity: 0, wantErr: ErrInvalidQuantity},
		{name: "quantity over max", giftID: "coche", quantity: MaxQuantity + 1, wantErr: ErrInvalidQuantity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNewReservation(tt.giftID, tt.quantity)
			if err != tt.wantErr {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestGiftAvailable(t *testing.T) {
	if got := (Gift{Required: 6, Selected: 2}).Available(); got != 4 {
		t.Fatalf("Available() = %d, want 4", got)
	}
	if got := (Gift{Required: 1, Selected: 3}).Available(); got != 0 {
		t.Fatalf("Available() = %d, want 0", got)
	}
}
