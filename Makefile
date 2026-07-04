BINARY_NAME := dra
GO := /usr/local/go/bin/go
GOFLAGS := -trimpath
LDFLAGS := -s -w -X dra/internal/cli.version=0.10.2-go

.PHONY: all build clean test install docker

all: build

build:
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .

install: build
	cp $(BINARY_NAME) /usr/local/bin/

test:
	$(GO) test ./...

clean:
	rm -f $(BINARY_NAME)

# Cross-compile for various platforms
.PHONY: build-linux-amd64 build-linux-arm64 build-linux-arm build-macos-amd64 build-macos-arm64

build-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-linux-amd64 .

build-linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-linux-arm64 .

build-linux-arm:
	GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-linux-arm .

build-macos-amd64:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-darwin-amd64 .

build-macos-arm64:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-darwin-arm64 .

# Build all platforms
build-all: build-linux-amd64 build-linux-arm64 build-linux-arm build-macos-amd64 build-macos-arm64


# Docker build
docker:
	docker build -t dra:latest .


# Run inside docker
docker-run:
	docker run --rm -it dra:latest download --automatic $(REPO)

# Build and push multi-arch image to registry
# Usage: make docker-push DOCKER_REGISTRY=docker.io/username DOCKER_TAG=latest
.PHONY:  docker-push

DOCKER_REGISTRY ?= your-registry.com/your-username
DOCKER_IMAGE ?= dra
DOCKER_TAG ?= latest

docker-push:
	docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v6 \
		--no-cache \
		--tag $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG) \
		--push .

