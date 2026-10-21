# syntax=docker/dockerfile:1
FROM golang:1.25.9-alpine AS builder

WORKDIR /app

ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /app/main ./cmd/api

# ===========================================
# Final Image
# ===========================================
FROM alpine:3.22

# Certificados raíz (SSL de Neon) y zona horaria
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -H appuser

WORKDIR /opt/app

COPY --from=builder /app/main .

USER appuser

EXPOSE 8080

ENV APP_NAME=regalos-lian-api
ENV SERVICE_PORT=8080
ENV DB_SSL_MODE=verify-full
ENV LOG_LEVEL=info

# DB_ADDR, DB_DATABASE, DB_USER, ALLOWED_ORIGINS: variables de Cloud Run
# DB_PASSWORD, ADMIN_TOKEN: secretos de Secret Manager

CMD ["/opt/app/main"]
