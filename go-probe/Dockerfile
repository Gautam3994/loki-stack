# ---------------------------------------------------------
# 1. BUILD STAGE (Optimized with BuildKit)
# ---------------------------------------------------------
FROM golang:1.26-alpine AS builder
WORKDIR /app

# Copy our source code and tests
COPY main.go main_test.go go.mod ./

# Run tests
RUN go test -v

# Build the binary using BuildKit cache mounts and dynamic architecture.
# $TARGETARCH ensures it cross-compiles correctly (e.g., arm64 for Mac, amd64 for EC2)
ARG TARGETARCH
ARG TARGETOS
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-s -w" -o healthcheck main.go

# ---------------------------------------------------------
# 2. RUNTIME STAGE (Grafana Loki)
# ---------------------------------------------------------
FROM grafana/loki:3.7.0

# Switch to root just long enough to copy the binary and make it executable
USER root
COPY --chmod=755 --from=builder /app/healthcheck /usr/local/bin/healthcheck

# Immediately switch back to Loki's default non-root user (UID 10001) for security
USER 10001

# We do NOT define an ENTRYPOINT. 
# Loki's base image will naturally start the Loki server as PID 1.