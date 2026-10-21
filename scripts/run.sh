#!/bin/sh
# Uso: ./scripts/run.sh dev
# Las variables se leen de .env (opcional) y .dev.env.

enviroment="$1"

if [ "$enviroment" = 'dev' ]
then
  echo "running in mode development"
  go run cmd/main.go
fi
