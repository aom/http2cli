.PHONY: build build-linux-amd64 build-linux-arm64 build-darwin-arm64 build-windows-amd64 build-all run test test-integration test-all clean docker docker-dev

# Build binary for current platform
build:
	go build -o http2cli ./cmd/http2cli

# Build for Linux amd64 (Intel/AMD)
build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o http2cli-linux-amd64 ./cmd/http2cli

# Build for Linux arm64 (AWS Graviton, etc.)
build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o http2cli-linux-arm64 ./cmd/http2cli

# Build for macOS arm64 (Apple Silicon)
build-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o http2cli-darwin-arm64 ./cmd/http2cli

# Build for Windows amd64
build-windows-amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o http2cli-windows-amd64.exe ./cmd/http2cli

# Build all platforms
build-all: build-linux-amd64 build-linux-arm64 build-darwin-arm64 build-windows-amd64

# Run locally
run: build
	./http2cli --config configs/config.yaml

# Run unit tests
test:
	go test ./...

# Run integration tests in Docker
test-integration: docker-dev
	docker run --rm http2cli:dev /tests/integration.sh

# Run all tests
test-all: test test-integration

# Clean build artifacts
clean:
	rm -f http2cli http2cli-linux-amd64 http2cli-linux-arm64 http2cli-darwin-arm64 http2cli-windows-amd64.exe

# Build production Docker image (binary only)
docker:
	docker build -t http2cli .

# Build development Docker image (with Alpine CLI tools)
docker-dev:
	docker build -f Dockerfile.dev -t http2cli:dev .

# Run development container
docker-run: docker-dev
	docker run --rm -p 8080:8080 http2cli:dev
