BINARY_NAME=ocserv-users-web-linux-amd64
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X 'ocserv-users/internal.Version=$(VERSION)' \
                     -X 'ocserv-users/internal.GitCommit=$(GIT_COMMIT)' \
                     -X 'ocserv-users/internal.BuildTime=$(BUILD_TIME)'"

.PHONY: build clean version lint lint-go lint-html fmt fmt-go fmt-html 

build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME)

version:
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"


# ========== Lint ==========

lint: lint-go lint-html
	@echo "✅ All lint checks passed"

lint-go:
	golangci-lint run ./...

lint-html:
	npx prettier --check web/templates/**/** && npx prettier --check web/templates/**

# ========== Format ==========

fmt: fmt-go fmt-html
	@echo "✅ All files formatted"

fmt-go:
	gofmt -w .
	goimports -local ocserv-users -w .

fmt-html:
	npx prettier --write web/templates/**/** && npx prettier --write web/templates/**
