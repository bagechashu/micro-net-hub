BINARY_NAME=ocserv-users-web-linux-amd64
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X 'ocserv-users/internal.Version=$(VERSION)' \
                     -X 'ocserv-users/internal.GitCommit=$(GIT_COMMIT)' \
                     -X 'ocserv-users/internal.BuildTime=$(BUILD_TIME)'"

.PHONY: build clean version

build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME)

version:
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"