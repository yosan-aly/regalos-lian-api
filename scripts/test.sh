#!/bin/sh
# Uso: ./scripts/test.sh unit [cover|coverHTML]
#      ./scripts/test.sh it      (requiere Postgres local en :5433; usa el schema babyshower_test)

typeTest="$1"
coverMode="$2"

if [ "$typeTest" = 'unit' ]
then
  echo "running unit test..."
  if [ "$coverMode" = 'coverHTML' ]
  then
    go test -count=1 -covermode=count -coverprofile=coverage $(go list ./internal/...); go tool cover -html=coverage -o coverage.html
  elif [ "$coverMode" = 'cover' ]
  then
    go test -count=1 -covermode=count -coverprofile=coverage $(go list ./internal/...)
  else
    go test -count=1 $(go list ./internal/...)
  fi
fi

if [ "$typeTest" = 'it' ]
then
  echo "running integration test..."
  TEST_DATABASE_URL="${TEST_DATABASE_URL:-postgres://postgres:postgres@localhost:5433/babyshower?sslmode=disable}" \
    go test -count=1 $(go list ./internal/...)
fi
