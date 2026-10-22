package rest

import (
	"time"

	"github.com/yosan-aly/regalos-lian-api/internal/gifts/domain"
)

// CatalogResponse es la lista completa de regalos agrupada por categoría.
type CatalogResponse struct {
	Categories []CategoryDTO `json:"categories"`
}

type CategoryDTO struct {
	ID          string    `json:"id" example:"ropita"`
	Title       string    `json:"title" example:"Ropita y accesorios"`
	Description string    `json:"description" example:"Prendas por talla, pañales, toallitas y accesorios para vestir a Lian."`
	Gifts       []GiftDTO `json:"gifts"`
}

type GiftDTO struct {
	ID          string `json:"id" example:"bodies-3-6"`
	Name        string `json:"name" example:"Bodies de algodón"`
	Variant     string `json:"variant,omitempty" example:"3–6 meses"`
	Description string `json:"description" example:"Bodies para uso diario; talla 3–6 meses."`
	Required    int    `json:"required" example:"6"`
	Selected    int    `json:"selected" example:"2"`
	Available   int    `json:"available" example:"4"`
}

// ReserveRequest es opcional: sin cuerpo se reserva 1 unidad.
type ReserveRequest struct {
	Quantity int `json:"quantity" example:"1" minimum:"1" maximum:"20"`
}

type ReservationResponse struct {
	ID        string    `json:"id" example:"8cf95ed5-1e97-4386-a331-73a1ab7bea9e"`
	Quantity  int       `json:"quantity" example:"1"`
	CreatedAt time.Time `json:"createdAt" example:"2026-12-01T15:04:05Z"`
	Gift      GiftDTO   `json:"gift"`
}

type AdminReservationsResponse struct {
	Reservations []AdminReservationDTO `json:"reservations"`
}

type AdminReservationDTO struct {
	ID          string     `json:"id" example:"8cf95ed5-1e97-4386-a331-73a1ab7bea9e"`
	GiftID      string     `json:"giftId" example:"bodies-3-6"`
	GiftName    string     `json:"giftName" example:"Bodies de algodón"`
	GiftVariant string     `json:"giftVariant,omitempty" example:"3–6 meses"`
	Quantity    int        `json:"quantity" example:"2"`
	CreatedAt   time.Time  `json:"createdAt" example:"2026-12-01T15:04:05Z"`
	CancelledAt *time.Time `json:"cancelledAt,omitempty"`
}

func toGiftDTO(g domain.Gift) GiftDTO {
	return GiftDTO{
		ID:          g.ID,
		Name:        g.Name,
		Variant:     g.Variant,
		Description: g.Description,
		Required:    g.Required,
		Selected:    g.Selected,
		Available:   g.Available(),
	}
}

func toCatalogResponse(categories []domain.Category) CatalogResponse {
	resp := CatalogResponse{Categories: make([]CategoryDTO, 0, len(categories))}
	for _, c := range categories {
		dto := CategoryDTO{ID: c.ID, Title: c.Title, Description: c.Description, Gifts: make([]GiftDTO, 0, len(c.Gifts))}
		for _, g := range c.Gifts {
			dto.Gifts = append(dto.Gifts, toGiftDTO(g))
		}
		resp.Categories = append(resp.Categories, dto)
	}
	return resp
}

func toAdminReservationDTO(r domain.Reservation) AdminReservationDTO {
	return AdminReservationDTO{
		ID:          r.ID,
		GiftID:      r.GiftID,
		GiftName:    r.GiftName,
		GiftVariant: r.GiftVariant,
		Quantity:    r.Quantity,
		CreatedAt:   r.CreatedAt,
		CancelledAt: r.CancelledAt,
	}
}
