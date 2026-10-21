# regalos-lian-api

Backend de la lista de regalos del baby shower de Lian. Los invitados eligen un regalo
desde la página ([regalos-lian](https://github.com/yosan-aly/regalos-lian)) y la
disponibilidad se actualiza para todos, **sin regalos repetidos**. La selección es
anónima: no se piden nombres ni datos de contacto.

## Stack

- **Go 1.25** · [Fiber v2](https://gofiber.io) · [pgx v5](https://github.com/jackc/pgx)
- **PostgreSQL 17** ([Neon](https://neon.tech) en producción)
- **Google Cloud Run** (escala a 0) · Cloud Build · Secret Manager
- Logs JSON con `log/slog`, compatibles con Cloud Logging

## Arquitectura

Hexagonal en un solo servicio: el dominio no conoce ni HTTP ni la base de datos.

```
cmd/api/main.go                       # Arranque: config → BD → migraciones → HTTP
internal/
├── gifts/
│   ├── domain/                       # Entidades, reglas y errores de negocio
│   │   ├── gift.go                   #   Gift, Category, Available()
│   │   ├── reservation.go            #   Validación de la reserva
│   │   ├── errors.go                 #   Errores con código estable para el front
│   │   └── repository.go             #   Puerto de salida (interfaz)
│   ├── application/service.go        # Casos de uso
│   └── infrastructure/
│       ├── postgres/                 # Adaptador de salida (pgx)
│       └── rest/                     # Adaptador de entrada (Fiber): handlers y DTOs
└── platform/
    ├── config/                       # Variables de entorno
    ├── database/                     # Pool de conexiones y migraciones embebidas
    ├── logger/                       # slog → formato Cloud Logging
    └── server/                       # Middlewares, rutas, autenticación de admin
```

### Cómo se evitan los regalos repetidos

La reserva es una sola transacción con un `UPDATE` condicional:

```sql
UPDATE gifts SET selected = selected + $2
WHERE id = $1 AND selected + $2 <= required
RETURNING ...;            -- 0 filas → 409 NOT_ENOUGH_AVAILABLE
INSERT INTO reservations ...;
```

El `UPDATE` bloquea la fila del regalo: si dos invitados reservan al mismo tiempo, el
segundo espera y reevalúa la condición con el valor ya actualizado. Además, un
`CHECK (selected BETWEEN 0 AND required)` en la tabla es la última barrera.
El test `TestReserve_ConcurrentGuestsNeverOverbook` lanza 20 reservas simultáneas
sobre un regalo de 3 unidades y verifica que solo 3 lo logren.

## API

Base: `/api/v1`. Los errores siempre tienen la forma
`{"error": {"code": "...", "message": "..."}}`, y `message` se puede mostrar al invitado.

| Método | Ruta | Acceso | Descripción |
|---|---|---|---|
| GET | `/health` | público | Estado del servicio |
| GET | `/api/v1/catalog` | público | Categorías y regalos con `required`, `selected` y `available` |
| POST | `/api/v1/gifts/{giftId}/reservations` | público | Reserva. Body opcional: `{"quantity": 2}` (por defecto 1) |
| GET | `/api/v1/admin/reservations` | admin | Reservas realizadas (`?includeCancelled=true`) |
| DELETE | `/api/v1/admin/reservations/{id}` | admin | Cancela la reserva y libera las unidades |

Respuestas de `POST .../reservations`:

| Código | `error.code` | Cuándo |
|---|---|---|
| 201 | — | Reserva creada; devuelve el regalo con la disponibilidad actualizada |
| 400 | `INVALID_QUANTITY`, `INVALID_BODY` | Datos inválidos |
| 404 | `GIFT_NOT_FOUND` | El regalo no existe |
| 409 | `NOT_ENOUGH_AVAILABLE` | Otro invitado lo eligió primero |
| 429 | `TOO_MANY_REQUESTS` | Límite de reservas por minuto por IP |

Las rutas de admin requieren el header `X-Admin-Token`. Si `ADMIN_TOKEN` no está
definido, esas rutas responden 404. Hay ejemplos listos en [`requests.http`](requests.http).

## Desarrollo local

Requisitos: Go 1.25 y PostgreSQL (por defecto en `localhost:5433`).

```bash
# 1. Crear las bases (una vez)
psql -h localhost -p 5433 -U postgres -c "CREATE DATABASE babyshower"
psql -h localhost -p 5433 -U postgres -c "CREATE DATABASE babyshower_test"

# 2. Ejecutar (lee .dev.env; para sobrescribir valores crea un .env)
go run ./cmd/api              # http://localhost:7070

# 3. Tests
go test ./...                 # unitarios
TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5433/babyshower_test?sslmode=disable" \
  go test -count=1 ./...      # incluye integración contra Postgres
```

Las migraciones (`internal/platform/database/migrations`) se aplican solas al arrancar,
en orden y una sola vez; un advisory lock evita que dos instancias migren a la vez.

## Configuración

| Variable | Requerida | Descripción |
|---|---|---|
| `DB_ADDR` | sí | `host:puerto` de PostgreSQL |
| `DB_DATABASE` | sí | Nombre de la base |
| `DB_USER` / `DB_PASSWORD` | sí | Credenciales |
| `DB_SSL_MODE` | no | `verify-full` en producción (Neon) |
| `ALLOWED_ORIGINS` | sí | Orígenes CORS separados por comas |
| `ADMIN_TOKEN` | no | Token de las rutas de admin |
| `RESERVE_RATE_LIMIT` | no | Reservas por minuto por IP (por defecto 10) |
| `SERVICE_PORT` / `PORT` | no | Puerto HTTP (por defecto 8080) |
| `LOG_LEVEL` | no | `debug`, `info`, `warn` o `error` |

## Despliegue

```bash
gcloud builds submit --config=cloudbuild.yaml --substitutions="_TAG=$(git rev-parse --short HEAD)"
```

Construye la imagen, la publica en Artifact Registry y la despliega en Cloud Run.
`DB_PASSWORD` y `ADMIN_TOKEN` se leen de Secret Manager.
