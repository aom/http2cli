# Stage 1: Build the binary
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install git for go mod (if needed)
RUN apk add --no-cache git

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
# CGO_ENABLED=0 creates statically linked binary
# -ldflags="-w -s" strips debug info for smaller size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o http2cli ./cmd/http2cli

# Stage 2: Minimal image with just the binary
FROM scratch

# Copy binary
COPY --from=builder /build/http2cli /http2cli

# Default entrypoint
ENTRYPOINT ["/http2cli"]
