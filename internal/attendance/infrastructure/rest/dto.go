package rest

import (
	"time"

	"github.com/yosan-aly/regalos-lian-api/internal/attendance/domain"
)

// ConfirmationRequest es el formulario de confirmación de asistencia.
type ConfirmationRequest struct {
	Name      string `json:"name" example:"Ana María"`
	Attendees int    `json:"attendees" example:"3" minimum:"1" maximum:"10"`
}

type ConfirmationResponse struct {
	ID        string    `json:"id" example:"8cf95ed5-1e97-4386-a331-73a1ab7bea9e"`
	Name      string    `json:"name" example:"Ana María"`
	Attendees int       `json:"attendees" example:"3"`
	CreatedAt time.Time `json:"createdAt" example:"2026-11-20T15:04:05Z"`
	UpdatedAt time.Time `json:"updatedAt" example:"2026-11-21T09:30:00Z"`
}

type ConfirmationSummaryResponse struct {
	Confirmations      []ConfirmationResponse `json:"confirmations"`
	TotalConfirmations int                    `json:"totalConfirmations" example:"12"`
	TotalAttendees     int                    `json:"totalAttendees" example:"31"`
}

func toConfirmationResponse(c domain.Confirmation) ConfirmationResponse {
	return ConfirmationResponse{
		ID:        c.ID,
		Name:      c.Name,
		Attendees: c.Attendees,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func toSummaryResponse(s domain.Summary) ConfirmationSummaryResponse {
	resp := ConfirmationSummaryResponse{
		Confirmations:      make([]ConfirmationResponse, 0, len(s.Confirmations)),
		TotalConfirmations: len(s.Confirmations),
		TotalAttendees:     s.TotalAttendees,
	}
	for _, c := range s.Confirmations {
		resp.Confirmations = append(resp.Confirmations, toConfirmationResponse(c))
	}
	return resp
}
