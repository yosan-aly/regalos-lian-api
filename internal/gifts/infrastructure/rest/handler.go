package rest

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/yosan-aly/regalos-lian-api/internal/gifts/domain"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/httperr"
)

// GiftService es lo que el adaptador HTTP necesita de la capa de aplicación.
type GiftService interface {
	Catalog(ctx context.Context) ([]domain.Category, error)
	Reserve(ctx context.Context, giftID string, quantity int) (domain.Reservation, domain.Gift, error)
	Reservations(ctx context.Context, includeCancelled bool) ([]domain.Reservation, error)
	CancelReservation(ctx context.Context, id string) error
}

type Handler struct {
	svc GiftService
}

func NewHandler(svc GiftService) *Handler {
	return &Handler{svc: svc}
}

// Catalog
// @Summary Catálogo de regalos
// @Description Devuelve las categorías con sus regalos y la disponibilidad actual de cada uno.
// @Tags Catalog
// @Produce application/json
// @Success 200 {object} CatalogResponse
// @Failure 500 {object} httperr.ErrorResponse
// @Router /catalog [get]
func (h *Handler) Catalog(c *fiber.Ctx) error {
	categories, err := h.svc.Catalog(c.UserContext())
	if err != nil {
		return err
	}
	// La disponibilidad cambia con cada reserva: nada de caché.
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(toCatalogResponse(categories))
}

// Reserve
// @Summary Reservar un regalo
// @Description Reserva unidades de un regalo de forma anónima y atómica.
// @Description El cuerpo es opcional: sin cuerpo se reserva 1 unidad.
// @Tags Reservations
// @Accept application/json
// @Produce application/json
// @Param giftId path string true "Id del regalo" default(bodies-3-6)
// @Param Reservation body ReserveRequest false "Cantidad a reservar (1 a 20)"
// @Success 201 {object} ReservationResponse
// @Failure 400 {object} httperr.ErrorResponse "INVALID_QUANTITY o INVALID_BODY"
// @Failure 404 {object} httperr.ErrorResponse "GIFT_NOT_FOUND"
// @Failure 409 {object} httperr.ErrorResponse "NOT_ENOUGH_AVAILABLE: otro invitado se adelantó"
// @Failure 429 {object} httperr.ErrorResponse "TOO_MANY_REQUESTS"
// @Failure 500 {object} httperr.ErrorResponse
// @Router /gifts/{giftId}/reservations [post]
func (h *Handler) Reserve(c *fiber.Ctx) error {
	var req ReserveRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return httperr.Write(c, fiber.StatusBadRequest, "INVALID_BODY", "El cuerpo de la solicitud no es válido.")
		}
	}
	if req.Quantity == 0 {
		req.Quantity = 1
	}

	res, gift, err := h.svc.Reserve(c.UserContext(), c.Params("giftId"), req.Quantity)
	if err != nil {
		return h.fail(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(ReservationResponse{
		ID:        res.ID,
		Quantity:  res.Quantity,
		CreatedAt: res.CreatedAt,
		Gift:      toGiftDTO(gift),
	})
}

// Reservations
// @Summary Listar reservas
// @Description Reservas realizadas, de la más reciente a la más antigua.
// @Tags Admin
// @Produce application/json
// @Security AdminToken
// @Param includeCancelled query bool false "Incluir reservas canceladas"
// @Success 200 {object} AdminReservationsResponse
// @Failure 401 {object} httperr.ErrorResponse "UNAUTHORIZED"
// @Failure 500 {object} httperr.ErrorResponse
// @Router /admin/reservations [get]
func (h *Handler) Reservations(c *fiber.Ctx) error {
	list, err := h.svc.Reservations(c.UserContext(), c.QueryBool("includeCancelled"))
	if err != nil {
		return err
	}
	resp := AdminReservationsResponse{Reservations: make([]AdminReservationDTO, 0, len(list))}
	for _, r := range list {
		resp.Reservations = append(resp.Reservations, toAdminReservationDTO(r))
	}
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(resp)
}

// CancelReservation
// @Summary Cancelar una reserva
// @Description Marca la reserva como cancelada y devuelve las unidades al regalo.
// @Tags Admin
// @Produce application/json
// @Security AdminToken
// @Param id path string true "Id de la reserva (UUID)"
// @Success 204 "Reserva cancelada"
// @Failure 401 {object} httperr.ErrorResponse "UNAUTHORIZED"
// @Failure 404 {object} httperr.ErrorResponse "RESERVATION_NOT_FOUND"
// @Failure 500 {object} httperr.ErrorResponse
// @Router /admin/reservations/{id} [delete]
func (h *Handler) CancelReservation(c *fiber.Ctx) error {
	if err := h.svc.CancelReservation(c.UserContext(), c.Params("id")); err != nil {
		return h.fail(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// fail traduce errores de negocio a respuestas HTTP; el resto sube al
// ErrorHandler global, que responde 500 y lo registra en el log.
func (h *Handler) fail(c *fiber.Ctx, err error) error {
	var de *domain.Error
	if !errors.As(err, &de) {
		return err
	}
	return httperr.Write(c, statusFor(de), de.Code, de.Message)
}

func statusFor(err *domain.Error) int {
	switch err {
	case domain.ErrGiftNotFound, domain.ErrReservationNotFound:
		return fiber.StatusNotFound
	case domain.ErrNotEnoughAvailable:
		return fiber.StatusConflict
	default:
		return fiber.StatusBadRequest
	}
}
