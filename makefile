# Makefile for compiling Golang Linux installation package and frontend package

# builder and cleaner
GO := go
GO_BUILD := $(GO) build
GO_CLEAN := $(GO) clean
GO_INSTALL := $(GO) install
NPM := npm

# cross-compiling Go for Linux
GOOS := linux
GOARCH := amd64

# Variables
WEB_SRC_DIR := frontend
SERVICE_SRC_DIR := backend
APP_NAME := micro-net-hub-$(GOOS)-$(GOARCH)
BIN_DIR := ../bin
GO_SRC := ./cmd/micro-net-hub/main.go

all: fe be

be:
	@echo "===== compile $(GOOS) $(GOARCH)... ====="
	cd $(SERVICE_SRC_DIR) && env GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO_BUILD) -o $(BIN_DIR)/$(APP_NAME) $(GO_SRC)
	@echo "target: $(BIN_DIR)/$(APP_NAME)"

fe:
	@echo "===== compile webui... ====="
	cd $(WEB_SRC_DIR) && npm run build:prod

clean:
	$(GO_CLEAN)
	rm -f $(SERVICE_SRC_DIR)/$(BIN_DIR)/$(APP_NAME)

.DEFAULT_GOAL := all

.PHONY: all be fe clean 
