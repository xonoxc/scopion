# =========================
# Project config
# =========================
APP_NAME := scopion
GO_CMD := go
UI_DIR := ui
UI_DIST := ui/dist
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP_NAME)

PORT ?= 8080
DEMO ?= true

# =========================
# Default target
# =========================
.DEFAULT_GOAL := help

# =========================
# Help
# =========================
help:
	@echo ""
	@echo "$(APP_NAME) – single-binary observability"
	@echo ""
	@echo "Usage:"
	@echo "  make dev          Run backend + UI dev server (NOT embedded)"
	@echo "  make ui-build     Build UI for embedding"
	@echo "  make build        Build production binary (with embedded UI)"
	@echo "  make run          Run embedded binary"
	@echo "  make test         Run all tests"
	@echo "  make clean        Clean build artifacts"
	@echo ""

# =========================
# UI
# =========================
ui-install:
	cd $(UI_DIR) && npm install

ui-dev:
	cd $(UI_DIR) && npm run dev

ui-build:
	@echo "Building UI for embedding..."
	cd $(UI_DIR) && npm run build
	@echo "UI build complete"

# =========================
# Backend
# =========================
run: build
	@$(BIN) start --port $(PORT) --demo=$(DEMO)

build: ui-build
	@echo "Building $(APP_NAME) binary..."
	mkdir -p $(BIN_DIR)
	$(GO_CMD) build -o $(BIN) ./cmd/$(APP_NAME)
	@echo "Built $(BIN)"

# =========================
# Dev (UI + Backend)
# =========================
dev:
	@echo "Starting UI dev server (http://localhost:5173)..."
	cd $(UI_DIR) && npm run dev & \
	echo "Starting Go backend (no embedded UI)..." && \
	$(GO_CMD) run ./cmd/$(APP_NAME) start --port $(PORT) --demo=$(DEMO)

# =========================
# Tests
# =========================
test:
	$(GO_CMD) test ./...
	cd clients/python && python -m unittest test_scopion_client.py
	cd clients/typescript && bun run test

test-race:
	$(GO_CMD) test -race ./...

# =========================
# Clean
# =========================
clean:
	rm -rf $(BIN_DIR)
	rm -rf $(UI_DIST)
	@echo "Cleaned build artifacts"
