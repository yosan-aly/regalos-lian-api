package rest

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/yosan-aly/regalos-lian-api/internal/attendance/domain"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/httperr"
)

// AttendanceService es lo que el adaptador HTTP necesita de la capa de aplicación.
type AttendanceService interface {
	Confirm(ctx context.Context, name string, attendees int) (domain.Confirmation, error)
	Update(ctx context.Context, id, name string, attendees int) (domain.Confirmation, error)
	Summary(ctx context.Context) (domain.Summary, error)
	Delete(ctx context.Context, id string) error
}

type Handler struct {
	svc AttendanceService
}

func NewHandler(svc AttendanceService) *Handler {
	return &Handler{svc: svc}
}

// Confirm
// @Summary Confirmar asistencia
// @Description Registra que un invitado asistirá y cuántas personas van en total (incluyéndolo).
// @Description Guarda el id de la respuesta: es el que permite actualizar la confirmación.
// @Tags Attendance
// @Accept application/json
// @Produce application/json
// @Param Confirmation body ConfirmationRequest true "Nombre y total de asistentes (1 a 7)"
// @Success 201 {object} ConfirmationResponse
// @Failure 400 {object} httperr.ErrorResponse "INVALID_NAME, INVALID_ATTENDEES o INVALID_BODY"
// @Failure 429 {object} httperr.ErrorResponse "TOO_MANY_REQUESTS"
// @Failure 500 {object} httperr.ErrorResponse
// @Router /attendance-confirmations [post]
func (h *Handler) Confirm(c *fiber.Ctx) error {
	var req ConfirmationRequest
	if err := c.BodyParser(&req); err != nil {
		return invalidBody(c)
	}
	conf, err := h.svc.Confirm(c.UserContext(), req.Name, req.Attendees)
	if err != nil {
		return h.fail(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(toConfirmationResponse(conf))
}

// Update
// @Summary Actualizar una confirmación
// @Description Cambia el nombre o la cantidad de asistentes. Requiere el id que se recibió al confirmar.
// @Tags Attendance
// @Accept application/json
// @Produce application/json
// @Param id path string true "Id de la confirmación (UUID)"
// @Param Confirmation body ConfirmationRequest true "Nombre y total de asistentes (1 a 7)"
// @Success 200 {object} ConfirmationResponse
// @Failure 400 {object} httperr.ErrorResponse "INVALID_NAME, INVALID_ATTENDEES o INVALID_BODY"
// @Failure 404 {object} httperr.ErrorResponse "CONFIRMATION_NOT_FOUND"
// @Failure 429 {object} httperr.ErrorResponse "TOO_MANY_REQUESTS"
// @Failure 500 {object} httperr.ErrorResponse
// @Router /attendance-confirmations/{id} [put]
func (h *Handler) Update(c *fiber.Ctx) error {
	var req ConfirmationRequest
	if err := c.BodyParser(&req); err != nil {
		return invalidBody(c)
	}
	conf, err := h.svc.Update(c.UserContext(), c.Params("id"), req.Name, req.Attendees)
	if err != nil {
		return h.fail(c, err)
	}
	return c.JSON(toConfirmationResponse(conf))
}

// Summary
// @Summary Listar confirmaciones
// @Description Confirmaciones de asistencia (de la más reciente a la más antigua) con el total de asistentes.
// @Tags Admin
// @Produce application/json
// @Security AdminToken
// @Success 200 {object} ConfirmationSummaryResponse
// @Failure 401 {object} httperr.ErrorResponse "UNAUTHORIZED"
// @Failure 500 {object} httperr.ErrorResponse
// @Router /admin/attendance-confirmations [get]
func (h *Handler) Summary(c *fiber.Ctx) error {
	summary, err := h.svc.Summary(c.UserContext())
	if err != nil {
		return err
	}
	c.Set(fiber.HeaderCacheControl, "no-store")
	return c.JSON(toSummaryResponse(summary))
}

// Delete
// @Summary Eliminar una confirmación
// @Description Elimina una confirmación, por ejemplo un duplicado.
// @Tags Admin
// @Produce application/json
// @Security AdminToken
// @Param id path string true "Id de la confirmación (UUID)"
// @Success 204 "Confirmación eliminada"
// @Failure 401 {object} httperr.ErrorResponse "UNAUTHORIZED"
// @Failure 404 {object} httperr.ErrorResponse "CONFIRMATION_NOT_FOUND"
// @Failure 500 {object} httperr.ErrorResponse
// @Router /admin/attendance-confirmations/{id} [delete]
func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.svc.Delete(c.UserContext(), c.Params("id")); err != nil {
		return h.fail(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func invalidBody(c *fiber.Ctx) error {
	return httperr.Write(c, fiber.StatusBadRequest, "INVALID_BODY", "El cuerpo de la solicitud no es válido.")
}

// fail traduce errores de negocio a respuestas HTTP; el resto sube al
// ErrorHandler global, que responde 500 y lo registra en el log.
func (h *Handler) fail(c *fiber.Ctx, err error) error {
	var de *domain.Error
	if !errors.As(err, &de) {
		return err
	}
	status := fiber.StatusBadRequest
	if de == domain.ErrConfirmationNotFound {
		status = fiber.StatusNotFound
	}
	return httperr.Write(c, status, de.Code, de.Message)
}
