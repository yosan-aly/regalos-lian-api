# regalos-lian-api

Backend de la lista de regalos del baby shower de Lian. Los invitados eligen un regalo
desde la página ([regalos-lian](https://github.com/yosan-aly/regalos-lian)) y la
disponibilidad se actualiza para todos, **sin regalos repetidos**. La selección es
anónima: no se piden nombres ni datos de contacto.

También registra la **confirmación de asistencia** al evento: nombre y cuántas personas
asisten en total.

## Stack

- **Go 1.25** · [Fiber v2](https://gofiber.io) · [pgx v5](https://github.com/jackc/pgx)
- **PostgreSQL 17** ([Neon](https://neon.tech) en producción)
- **Google Cloud Run** (escala a 0) · Cloud Build · Secret Manager
- Logs JSON con `log/slog`, compatibles con Cloud Logging

## Arquitectura

Hexagonal en un solo servicio: el dominio no conoce ni HTTP ni la base de datos.

```
api/rest/                             # Especificación OpenAPI generada (swag)
cmd/main.go                           # Arranque: config → BD → migraciones → HTTP
scripts/                              # open-api.sh, run.sh, test.sh
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
├── attendance/                       # Confirmación de asistencia (misma estructura)
│   ├── domain/
│   ├── application/
│   └── infrastructure/{postgres,rest}/
└── platform/
    ├── config/                       # Variables de entorno
    ├── database/                     # Pool de conexiones, migraciones embebidas y dbtest
    ├── httperr/                      # Formato común de errores
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
| POST | `/api/v1/attendance-confirmations` | público | Confirma asistencia. Body: `{"name": "Ana", "attendees": 3}`; devuelve el `id` |
| PUT | `/api/v1/attendance-confirmations/{id}` | público, con el `id` | Actualiza nombre o asistentes |
| GET | `/api/v1/admin/attendance-confirmations` | admin | Confirmaciones con `totalConfirmations` y `totalAttendees` |
| DELETE | `/api/v1/admin/attendance-confirmations/{id}` | admin | Elimina una confirmación (por ejemplo, un duplicado) |

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

### Swagger y Postman

La especificación OpenAPI está en [`api/rest/swagger.json`](api/rest/swagger.json) y se
genera a partir de las anotaciones de los handlers:

```bash
sh scripts/open-api.sh        # tras cambiar endpoints o DTOs
```

- **Swagger UI:** con `SERVICE_WITH_OPEN_API=true` (activo en `.dev.env`), abre
  <http://localhost:7070/swagger/index.html>. En producción está desactivado.
- **Postman:** *Import* → archivo `api/rest/swagger.json` (o la URL
  `http://localhost:7070/swagger/doc.json`). Crea la colección con todos los endpoints;
  para los de admin, en la pestaña *Authorization* de la colección pon el valor de
  `X-Admin-Token` (`dev-admin-token` en local).

## Desarrollo local

Requisitos: Go 1.25 y PostgreSQL (por defecto en `localhost:5433`).

```bash
# 1. Crear la base (una vez); los schemas los crea el servicio
psql -h localhost -p 5433 -U postgres -c "CREATE DATABASE babyshower"

# 2. Ejecutar (lee .dev.env; para sobrescribir valores crea un .env)
go run ./cmd                  # http://localhost:7070

# 3. Tests
go test ./...                 # unitarios
TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5433/babyshower?sslmode=disable" \
  go test -count=1 ./...      # integración, en el schema babyshower_test
```

Las migraciones (`internal/platform/database/migrations`) se aplican solas al arrancar,
en orden y una sola vez. Si el schema de `DB_SCHEMAS` no existe, se crea; cada conexión
trabaja dentro de él (`search_path`). En local, la base `babyshower` tiene el schema
`babyshower` para desarrollo y `babyshower_test` para los tests de integración. Un advisory lock por schema evita que dos
instancias migren a la vez.

## Configuración

| Variable | Requerida | Descripción |
|---|---|---|
| `DB_ADDR` | sí | `host:puerto` de PostgreSQL |
| `DB_DATABASE` | sí | Nombre de la base |
| `DB_SCHEMAS` | no | Schema del servicio; se crea si no existe (por defecto `public`) |
| `DB_USER` / `DB_PASSWORD` | sí | Credenciales |
| `DB_SSL_MODE` | no | `verify-full` en producción (Neon) |
| `ALLOWED_ORIGINS` | sí | Orígenes CORS separados por comas |
| `ADMIN_TOKEN` | no | Token de las rutas de admin |
| `RESERVE_RATE_LIMIT` | no | Escrituras públicas (reservas y confirmaciones) por minuto por IP (por defecto 10) |
| `SERVICE_PORT` / `PORT` | no | Puerto HTTP (por defecto 8080) |
| `SERVICE_WITH_OPEN_API` | no | `true` publica Swagger UI en `/swagger` (por defecto `false`) |
| `LOG_LEVEL` | no | `debug`, `info`, `warn` o `error` |

## Despliegue

```bash
gcloud builds submit --config=cloudbuild.yaml --substitutions="_TAG=$(git rev-parse --short HEAD)"
```

Construye la imagen, la publica en Artifact Registry y la despliega en Cloud Run.
`DB_PASSWORD` y `ADMIN_TOKEN` se leen de Secret Manager.
