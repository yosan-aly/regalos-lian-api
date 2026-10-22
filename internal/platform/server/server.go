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
	"github.com/gofiber/swagger"

	_ "github.com/yosan-aly/regalos-lian-api/api/rest" // registra la especificación OpenAPI
	attendancerest "github.com/yosan-aly/regalos-lian-api/internal/attendance/infrastructure/rest"
	giftsrest "github.com/yosan-aly/regalos-lian-api/internal/gifts/infrastructure/rest"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/config"
	"github.com/yosan-aly/regalos-lian-api/internal/platform/httperr"
)

const adminTokenHeader = "X-Admin-Token"

// Services agrupa los casos de uso que expone la API.
type Services struct {
	Gifts      giftsrest.GiftService
	Attendance attendancerest.AttendanceService
}

// New arma la aplicación HTTP con sus middlewares y rutas.
func New(cfg config.Config, svc Services, log *slog.Logger) *fiber.App {
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
			AllowMethods: "GET,POST,PUT,DELETE",
			AllowHeaders: "Content-Type," + adminTokenHeader,
		}),
	)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	if cfg.WithOpenAPI {
		// La especificación se genera con scripts/open-api.sh en api/rest.
		app.Get("/swagger/*", swagger.HandlerDefault)
		log.Info("swagger enabled", "url", "http://localhost:"+cfg.Port+"/swagger/index.html")
	}

	gifts := giftsrest.NewHandler(svc.Gifts)
	attendance := attendancerest.NewHandler(svc.Attendance)
	// Un solo límite para las escrituras públicas: el contador por IP se comparte.
	writeLimit := writeLimiter(cfg.ReserveRateLimit)

	api := app.Group("/api/v1")
	api.Get("/catalog", gifts.Catalog)
	api.Post("/gifts/:giftId/reservations", writeLimit, gifts.Reserve)
	api.Post("/attendance-confirmations", writeLimit, attendance.Confirm)
	api.Put("/attendance-confirmations/:id", writeLimit, attendance.Update)

	admin := api.Group("/admin", adminAuth(cfg.AdminToken))
	admin.Get("/reservations", gifts.Reservations)
	admin.Delete("/reservations/:id", gifts.CancelReservation)
	admin.Get("/attendance-confirmations", attendance.Summary)
	admin.Delete("/attendance-confirmations/:id", attendance.Delete)

	return app
}

// writeLimiter evita que un script llene la lista de reservas o confirmaciones falsas.
func writeLimiter(perMinute int) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        perMinute,
		Expiration: time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return httperr.Write(c, fiber.StatusTooManyRequests, "TOO_MANY_REQUESTS",
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
			return httperr.Write(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Token de administración inválido.")
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
			return httperr.Write(c, fe.Code, "HTTP_ERROR", fe.Message)
		}
		log.Error("unhandled error",
			"error", err.Error(),
			"method", c.Method(),
			"path", c.Path(),
			"requestId", c.Locals(requestid.ConfigDefault.ContextKey),
		)
		return httperr.Write(c, fiber.StatusInternalServerError, "INTERNAL_ERROR",
			"Ocurrió un error inesperado. Inténtalo de nuevo en un momento.")
	}
}
