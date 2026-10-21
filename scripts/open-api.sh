#!/bin/sh

swag init -g './cmd/main.go' --parseDependency --useStructName -o './api/rest'
