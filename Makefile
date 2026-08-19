# =========================
# Project configuration
# =========================
APP_NAME := Hyper
BIN_DIR  := bin

GOOS    ?= linux
GOARCH  ?= amd64
CGO     ?= 0

GO   := go
WIRE := wire

# =========================
# Commands
# =========================
API_CMD    := ./cmd/api-server
CONN_CMD   := ./cmd/conn-server
FANOUT_CMD := ./cmd/fanout-server

# =========================
# Default
# =========================
.PHONY: all
all: gen build-api

# =========================
# Wire
# =========================
.PHONY: gen gen-api gen-conn gen-fanout

gen: gen-api gen-conn gen-fanout

gen-api:
	@echo "==> wire api-server"
	@cd $(API_CMD) && $(WIRE)

gen-conn:
	@echo "==> wire conn-server"
	@cd $(CONN_CMD) && $(WIRE)

#gen-fanout:
#	@echo "==> wire fanout-server"
#	@cd $(FANOUT_CMD) && $(WIRE)

.PHONY: build
build: build-api build-conn

# =========================
# Build
# =========================
.PHONY: build-api build-conn

build-api: gen-api
	@echo "==> build api-server"
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=$(CGO) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		  $(GO) build -o $(BIN_DIR)/api-server ./$(API_CMD)

build-conn: gen-conn
	@echo "==> build conn-server"
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=$(CGO) GOOS=$(GOOS) GOARCH=$(GOARCH) \
	 	$(GO)  build -o $(BIN_DIR)/conn-server $(CONN_CMD)/.

build-fanout: gen-fanout
	@echo "==> build fanout-server"
	@mkdir -p $(BIN_DIR)
	@CGO_ENABLED=$(CGO) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		$(GO) build -o $(BIN_DIR)/fanout-server ./$(FANOUT_CMD)

# =========================
# Run (dev)
# =========================
.PHONY: run-api run-conn

run-api: gen-api
	@$(GO) run ./$(API_CMD) serve

run-conn: gen-conn
	@$(GO) run ./$(CONN_CMD)/.

run-fanout: gen-fanout
	@cd $(FANOUT_CMD) && $(GO) run .

# =========================
# Test
# =========================
.PHONY: test
test:
	@$(GO) test ./... -race

# =========================
# Push (build + upload to server)
# =========================
SERVER_HOST ?= 8.156.86.220
SERVER_USER ?= root
SERVER_PASS ?= luoxiaorui0!
SERVER_DIR  ?= /root

.PHONY: push
push: build
	@echo "==> upload api-server & conn-server to $(SERVER_USER)@$(SERVER_HOST):$(SERVER_DIR)"
	@sshpass -p '$(SERVER_PASS)' scp -o StrictHostKeyChecking=no \
		$(BIN_DIR)/api-server $(SERVER_USER)@$(SERVER_HOST):$(SERVER_DIR)/.api-server.new
	@sshpass -p '$(SERVER_PASS)' scp -o StrictHostKeyChecking=no \
		$(BIN_DIR)/conn-server $(SERVER_USER)@$(SERVER_HOST):$(SERVER_DIR)/.conn-server.new
	@sshpass -p '$(SERVER_PASS)' ssh -o StrictHostKeyChecking=no $(SERVER_USER)@$(SERVER_HOST) \
		"cd $(SERVER_DIR) && chmod +x .api-server.new .conn-server.new && mv -f .api-server.new api-server && mv -f .conn-server.new conn-server"
	@echo "==> push done"

# =========================
# Clean
# =========================
.PHONY: clean
clean:
	@rm -rf $(BIN_DIR)
