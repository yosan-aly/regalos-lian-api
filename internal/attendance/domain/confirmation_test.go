package domain

import (
	"strings"
	"testing"
)

func TestParseConfirmationInput(t *testing.T) {
	tests := []struct {
		name      string
		guest     string
		attendees int
		want      ConfirmationInput
		wantErr   error
	}{
		{name: "valid", guest: "Ana", attendees: 3, want: ConfirmationInput{Name: "Ana", Attendees: 3}},
		{name: "normalizes spaces", guest: "  Sanly   Ortiz ", attendees: 1, want: ConfirmationInput{Name: "Sanly Ortiz", Attendees: 1}},
		{name: "accented name counts runes", guest: "Íñ", attendees: 1, want: ConfirmationInput{Name: "Íñ", Attendees: 1}},
		{name: "short name", guest: " A ", attendees: 1, wantErr: ErrInvalidName},
		{name: "long name", guest: strings.Repeat("a", MaxNameLength+1), attendees: 1, wantErr: ErrInvalidName},
		{name: "zero attendees", guest: "Ana", attendees: 0, wantErr: ErrInvalidAttendees},
		{name: "too many attendees", guest: "Ana", attendees: MaxAttendees + 1, wantErr: ErrInvalidAttendees},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConfirmationInput(tt.guest, tt.attendees)
			if err != tt.wantErr {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
