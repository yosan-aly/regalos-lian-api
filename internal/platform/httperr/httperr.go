// Package httperr define el formato común de error de la API.
package httperr

import "github.com/gofiber/fiber/v2"

// ErrorResponse es el cuerpo de todas las respuestas de error.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string `json:"code" example:"NOT_ENOUGH_AVAILABLE"`
	Message string `json:"message" example:"Ya no quedan suficientes unidades de este regalo. Actualiza la página para ver la disponibilidad."`
}

// Write responde con el formato de error común de la API.
func Write(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(ErrorResponse{Error: ErrorBody{Code: code, Message: message}})
}
