package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Config struct {
	AppName string
	Port    string
	DB      DB
	// AllowedOrigins es la lista de orígenes CORS separada por comas.
	AllowedOrigins string
	// AdminToken protege /api/v1/admin. Si está vacío, esas rutas se desactivan.
	AdminToken string
	// ReserveRateLimit es el máximo de reservas por minuto por IP.
	ReserveRateLimit int
	// WithOpenAPI publica Swagger UI en /swagger (solo para desarrollo).
	WithOpenAPI bool
	LogLevel    slog.Level
}

type DB struct {
	Addr     string
	Database string
	// Schema donde viven las tablas del servicio; se crea si no existe.
	// Permite compartir una base de datos con otros proyectos.
	Schema   string
	User     string
	Password string
	SSLMode  string
}

var schemaNamePattern = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)

// URL arma la cadena de conexión escapando usuario y contraseña.
func (d DB) URL() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(d.User, d.Password),
		Host:   d.Addr,
		Path:   "/" + d.Database,
	}
	if d.SSLMode != "" {
		u.RawQuery = url.Values{"sslmode": {d.SSLMode}}.Encode()
	}
	return u.String()
}

// Load lee la configuración de variables de entorno.
func Load() (Config, error) {
	var missing []string
	required := func(key string) string {
		v := strings.TrimSpace(os.Getenv(key))
		if v == "" {
			missing = append(missing, key)
		}
		return v
	}

	cfg := Config{
		AppName: getEnv("APP_NAME", "regalos-lian-api"),
		// SERVICE_PORT sigue la convención de los otros servicios; PORT la define Cloud Run.
		Port: getEnv("SERVICE_PORT", getEnv("PORT", "8080")),
		DB: DB{
			Addr:     required("DB_ADDR"),
			Database: required("DB_DATABASE"),
			Schema:   getEnv("DB_SCHEMAS", "public"),
			User:     required("DB_USER"),
			Password: required("DB_PASSWORD"),
			SSLMode:  os.Getenv("DB_SSL_MODE"),
		},
		AllowedOrigins: required("ALLOWED_ORIGINS"),
		AdminToken:     os.Getenv("ADMIN_TOKEN"),
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("faltan variables de entorno: %s", strings.Join(missing, ", "))
	}

	if !schemaNamePattern.MatchString(cfg.DB.Schema) {
		return Config{}, errors.New("DB_SCHEMAS debe ser un solo schema en minúsculas (letras, números y _)")
	}

	limit, err := strconv.Atoi(getEnv("RESERVE_RATE_LIMIT", "10"))
	if err != nil || limit < 1 {
		return Config{}, errors.New("RESERVE_RATE_LIMIT debe ser un entero mayor que 0")
	}
	cfg.ReserveRateLimit = limit

	if cfg.WithOpenAPI, err = strconv.ParseBool(getEnv("SERVICE_WITH_OPEN_API", "false")); err != nil {
		return Config{}, errors.New("SERVICE_WITH_OPEN_API debe ser true o false")
	}

	if err := cfg.LogLevel.UnmarshalText([]byte(getEnv("LOG_LEVEL", "info"))); err != nil {
		return Config{}, fmt.Errorf("LOG_LEVEL inválido: %w", err)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
