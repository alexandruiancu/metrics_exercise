# Multi-stage Dockerfile for loader module (ZMQ consumer + OTel)
# Optimized for low-latency batch processing with OpenTelemetry SDK

# Stage 1: Builder
FROM golang:1.20-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum first for cache optimization
COPY go.work go.work.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the loader binary
# CGO_ENABLED=0 for static linking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo \
    -ldflags="-s -w -extldflags '-static'" \
    -o /loader .

# Stage 2: Runtime
FROM alpine:3.19 AS runtime

# Install runtime dependencies only
# ca-certificates for TLS/SSL support
RUN apk add --no-cache ca-certificates

# Non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /loader /app/loader

# Create directories for data and logs
RUN mkdir -p /app/data /app/logs && \
    chown -R appuser:appuser /app/data /app/logs

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9092/health || exit 1

# Expose ZMQ port (if applicable)
# EXPOSE 9092

# Run the loader binary
CMD ["/app/loader"]
