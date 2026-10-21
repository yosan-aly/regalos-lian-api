package rest

import (
	"time"

	"github.com/yosan-aly/regalos-lian-api/internal/gifts/domain"
)

type catalogResponse struct {
	Categories []categoryDTO `json:"categories"`
}

type categoryDTO struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Gifts       []giftDTO `json:"gifts"`
}

type giftDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Variant     string `json:"variant,omitempty"`
	Description string `json:"description"`
	Required    int    `json:"required"`
	Selected    int    `json:"selected"`
	Available   int    `json:"available"`
}

type reserveRequest struct {
	Quantity int `json:"quantity"`
}

type reservationResponse struct {
	ID        string    `json:"id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"createdAt"`
	Gift      giftDTO   `json:"gift"`
}

type adminReservationDTO struct {
	ID          string     `json:"id"`
	GiftID      string     `json:"giftId"`
	GiftName    string     `json:"giftName"`
	GiftVariant string     `json:"giftVariant,omitempty"`
	Quantity    int        `json:"quantity"`
	CreatedAt   time.Time  `json:"createdAt"`
	CancelledAt *time.Time `json:"cancelledAt,omitempty"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func toGiftDTO(g domain.Gift) giftDTO {
	return giftDTO{
		ID:          g.ID,
		Name:        g.Name,
		Variant:     g.Variant,
		Description: g.Description,
		Required:    g.Required,
		Selected:    g.Selected,
		Available:   g.Available(),
	}
}

func toCatalogResponse(categories []domain.Category) catalogResponse {
	resp := catalogResponse{Categories: make([]categoryDTO, 0, len(categories))}
	for _, c := range categories {
		dto := categoryDTO{ID: c.ID, Title: c.Title, Description: c.Description, Gifts: make([]giftDTO, 0, len(c.Gifts))}
		for _, g := range c.Gifts {
			dto.Gifts = append(dto.Gifts, toGiftDTO(g))
		}
		resp.Categories = append(resp.Categories, dto)
	}
	return resp
}

func toAdminReservationDTO(r domain.Reservation) adminReservationDTO {
	return adminReservationDTO{
		ID:          r.ID,
		GiftID:      r.GiftID,
		GiftName:    r.GiftName,
		GiftVariant: r.GiftVariant,
		Quantity:    r.Quantity,
		CreatedAt:   r.CreatedAt,
		CancelledAt: r.CancelledAt,
	}
}
