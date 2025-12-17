# =========================
# Project configuration
# =========================
APP_NAME := Hyper
CMD_DIR  := cmd/api-server
BIN_DIR  := bin

GO       := go
WIRE     := wire

GOFLAGS  :=
LDFLAGS  :=

# =========================
# Default target
# =========================
.PHONY: all
all: gen build

# =========================
# Code generation
# =========================
.PHONY: gen
gen: wire

.PHONY: wire
wire:
	@echo "==> Running wire"
	@cd $(CMD_DIR) && $(WIRE)

# =========================
# Build
# =========================
.PHONY: build
build:
	@echo "==> Building $(APP_NAME)"
	@mkdir -p $(BIN_DIR)
	@$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME) ./$(CMD_DIR)

# =========================
# Run
# =========================
.PHONY: run
run: gen
	@echo "==> Running $(APP_NAME)"
	@$(GO) run ./$(CMD_DIR) serve

# =========================
# Test
# =========================
.PHONY: test
test:
	@echo "==> Running tests"
	@$(GO) test ./... -race

# =========================
# Lint (optional)
# =========================
.PHONY: lint
lint:
	@echo "==> Running golangci-lint"
	@golangci-lint run

# =========================
# Clean
# =========================
.PHONY: clean
clean:
	@echo "==> Cleaning build artifacts"
	@rm -rf $(BIN_DIR)

# =========================
# Help
# =========================
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make gen     - Run code generation (wire)"
	@echo "  make build   - Build binary"
	@echo "  make run     - Generate + run"
	@echo "  make test    - Run all tests"
	@echo "  make lint    - Run golangci-lint"
	@echo "  make clean   - Remove build artifacts"
