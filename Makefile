TEST_DATABASE_URL ?= postgres://postgres:postgres@localhost:5433/babyshower_test?sslmode=disable

.PHONY: run test test-integration vet build docker-build

run:
	go run ./cmd/api

test:
	go test ./...

test-integration:
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -count=1 ./...

vet:
	go vet ./...

build:
	go build -trimpath -ldflags="-s -w" -o bin/regalos-lian-api ./cmd/api

docker-build:
	docker build -t regalos-lian-api:local .
