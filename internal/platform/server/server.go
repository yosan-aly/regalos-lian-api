package server

import (
	"crypto/subtle"
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/yosan-aly/regalos-lian-api/internal/gifts/infrastructure/rest"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/config"
)

const adminTokenHeader = "X-Admin-Token"

// New arma la aplicación HTTP con sus middlewares y rutas.
func New(cfg config.Config, svc rest.GiftService, log *slog.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               cfg.AppName,
		DisableStartupMessage: true,
		BodyLimit:             16 * 1024,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          15 * time.Second,
		IdleTimeout:           60 * time.Second,
		// Cloud Run pone la IP real del cliente en X-Forwarded-For;
		// el rate limit por IP depende de esto.
		ProxyHeader:        fiber.HeaderXForwardedFor,
		EnableIPValidation: true,
		ErrorHandler:       errorHandler(log),
	})

	app.Use(
		recover.New(),
		requestid.New(),
		requestLogger(log),
		cors.New(cors.Config{
			AllowOrigins: cfg.AllowedOrigins,
			AllowMethods: "GET,POST,DELETE",
			AllowHeaders: "Content-Type," + adminTokenHeader,
		}),
	)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	h := rest.NewHandler(svc)
	api := app.Group("/api/v1")
	api.Get("/catalog", h.Catalog)
	api.Post("/gifts/:giftId/reservations", reserveLimiter(cfg.ReserveRateLimit), h.Reserve)

	admin := api.Group("/admin", adminAuth(cfg.AdminToken))
	admin.Get("/reservations", h.Reservations)
	admin.Delete("/reservations/:id", h.CancelReservation)

	return app
}

// reserveLimiter evita que un script llene la lista de reservas falsas.
func reserveLimiter(perMinute int) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        perMinute,
		Expiration: time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return rest.WriteError(c, fiber.StatusTooManyRequests, "TOO_MANY_REQUESTS",
				"Demasiados intentos. Espera un minuto e inténtalo de nuevo.")
		},
	})
}

func adminAuth(token string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if token == "" {
			// Sin token configurado, las rutas de admin no existen.
			return fiber.ErrNotFound
		}
		got := c.Get(adminTokenHeader)
		if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			return rest.WriteError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Token de administración inválido.")
		}
		return c.Next()
	}
}

func requestLogger(log *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		if c.Path() == "/health" {
			return err
		}

		status := c.Response().StatusCode()
		var fe *fiber.Error
		if errors.As(err, &fe) {
			status = fe.Code
		} else if err != nil {
			status = fiber.StatusInternalServerError
		}

		log.Info("request",
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"latencyMs", time.Since(start).Milliseconds(),
			"requestId", c.Locals(requestid.ConfigDefault.ContextKey),
			"ip", c.IP(),
		)
		return err
	}
}

// errorHandler responde los errores no controlados sin filtrar detalles
// internos al cliente; el detalle queda solo en el log.
func errorHandler(log *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		var fe *fiber.Error
		if errors.As(err, &fe) {
			return rest.WriteError(c, fe.Code, "HTTP_ERROR", fe.Message)
		}
		log.Error("unhandled error",
			"error", err.Error(),
			"method", c.Method(),
			"path", c.Path(),
			"requestId", c.Locals(requestid.ConfigDefault.ContextKey),
		)
		return rest.WriteError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR",
			"Ocurrió un error inesperado. Inténtalo de nuevo en un momento.")
	}
}
