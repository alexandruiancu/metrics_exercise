# Multi-stage Dockerfile for Prometheus
# Optimized for low-latency metrics scraping

# Stage 1: Builder
FROM golang:1.20-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Download Prometheus binary
RUN wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz && \
    tar xzf prometheus-2.45.0.linux-amd64.tar.gz && \
    mv prometheus-2.45.0.linux-amd64/prometheus /prometheus && \
    rm -rf prometheus-2.45.0.linux-amd64.tar.gz prometheus-2.45.0.linux-amd64

# Stage 2: Runtime
FROM alpine:3.19 AS runtime

# Install runtime dependencies only
RUN apk add --no-cache ca-certificates

# Non-root user for security
RUN addgroup -g 2000 prometheus && \
    adduser -D -u 2000 -G prometheus prometheus && \
    chown -R prometheus:prometheus /app

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /prometheus /app/prometheus

# Create directories for data and config
RUN mkdir -p /app/data /app/config && \
    chown -R prometheus:prometheus /app/data /app/config

# Copy Prometheus configuration
COPY --from=builder /build/prometheus-2.45.0.linux-amd64/prometheus.yml /app/config/prometheus.yml
RUN chown prometheus:prometheus /app/config/prometheus.yml

# Switch to non-root user
USER prometheus

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9090/-/healthy || exit 1

# Expose metrics port
EXPOSE 9090

# Expose config reload endpoint
EXPOSE 9091

# Run Prometheus
CMD ["/app/prometheus", "--config.file=/app/config/prometheus.yml", "--storage.tsdb.path=/app/data"]
