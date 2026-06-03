.PHONY: build test lint clean release-snapshot docker install

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X github.com/replay/replay/internal/version.Version=$(VERSION) -X github.com/replay/replay/internal/version.Commit=$(COMMIT) -X github.com/replay/replay/internal/version.Date=$(DATE)

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o build/replay .

install:
	go install ./...

test:
	go test ./... -count=1

test-race:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w .

vet:
	go vet ./...

clean:
	rm -rf build/
	docker rmi replay:$(VERSION) 2>/dev/null || true

docker:
	docker build -t replay:$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg DATE=$(DATE) .

release-snapshot:
	goreleaser release --snapshot --clean

release-check:
	goreleaser check

deps:
	go mod tidy
	go mod verify

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build            Build the binary"
	@echo "  test             Run all tests"
	@echo "  test-race        Run tests with race detector"
	@echo "  lint             Run golangci-lint"
	@echo "  fmt              Format code"
	@echo "  vet              Run go vet"
	@echo "  clean            Remove build artifacts"
	@echo "  docker           Build Docker image"
	@echo "  release-snapshot Run GoReleaser snapshot"
	@echo "  release-check    Validate GoReleaser config"
	@echo "  deps             Tidy and verify dependencies"
	@echo "  install          Build and install replay binary via go install"
	@echo "  help             Show this help"
