### --- Makefile: db migrations (golang-migrate) --- ###

# Load .env if present (expects KEY=VALUE lines)
ifneq (,$(wildcard .env))
include .env
export
endif

# Defaults (overridden by .env or CLI VAR=val)
DB_USER ?= root
DB_PASSWORD ?= secret
DB_HOST ?= 127.0.0.1
DB_PORT ?= 3306
DB_NAME   ?= recrutr

# If you define MYSQL_DSN in .env, it will be used verbatim.
# Example: mysql://user:pass@tcp(127.0.0.1:3306)/authdb?multiStatements=true&parseTime=true
ifdef MYSQL_DSN
  DB_URL := $(MYSQL_DSN)
else
  DB_URL := mysql://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)?multiStatements=true&parseTime=true
endif

# Migrations path
MIGRATIONS_DIR ?= ./internal/db/migrations

# migrate binary (ensure on PATH, or use migrate-install)
MIGRATE_BIN ?= migrate

# Steps to go down (override with: make migrate-down STEPS=2)
STEPS ?= 1

# Target version for force operations (required for migrate-force)
VERSION ?=

ifeq ($(filter migrate-force,$(MAKECMDGOALS)),migrate-force)
ifeq ($(strip $(VERSION)),)
$(error VERSION is required. Usage: make migrate-force VERSION=2)
endif
endif

.PHONY: migrate-up migrate-down migrate-version migrate-install env-print migrate-force

## Apply all up migrations
migrate-up:
	@echo "==> UP migrations to $(DB_NAME) using $(MIGRATIONS_DIR)"
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" -verbose up

## Roll back N steps (default 1)
migrate-down:
	@echo "==> DOWN $(STEPS) step(s) on $(DB_NAME)"
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" -verbose down $(STEPS)

## Show current migration version
migrate-version:
	@$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" -verbose version

## Force database to a specific migration version (e.g., make migrate-force VERSION=2)
migrate-force:
	$(MIGRATE_BIN) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" -verbose force $(VERSION)

## Install golang-migrate CLI
migrate-install:
	GO111MODULE=on go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Installed: $$(go env GOPATH)/bin/migrate (add to PATH)"

## Debug: print resolved env (password not printed)
env-print:
	@echo "DB_HOST=$(DB_HOST)"
	@echo "DB_PORT=$(DB_PORT)"
	@echo "DB_NAME=$(DB_NAME)"
	@echo "DB_USER=$(DB_USER)"
	@echo "DB_PASSWORD=$(if $(filter secret,$(DB_PASSWORD)),****,$(DB_PASSWORD))"
	@echo "MIGRATIONS_DIR=$(MIGRATIONS_DIR)"
	@echo "Using MYSQL_DSN? $${MYSQL_DSN:+yes}{${MYSQL_DSN:=""}}"
