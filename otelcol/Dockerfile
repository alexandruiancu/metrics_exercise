# Multi-stage Dockerfile for OpenTelemetry Collector
# Optimized for low-latency metrics collection and export

# Stage 1: Builder
FROM golang:1.20-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Download OTel Collector binary
RUN wget https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.86.0/otelcol-contrib_0.86.0_linux_amd64.tar.gz && \
    tar xzf otelcol-contrib_0.86.0_linux_amd64.tar.gz && \
    mv otelcol-contrib_0.86.0_linux_amd64/otelcol-contrib /otelcol && \
    rm -rf otelcol-contrib_0.86.0_linux_amd64.tar.gz otelcol-contrib_0.86.0_linux_amd64

# Stage 2: Runtime
FROM alpine:3.19 AS runtime

# Install runtime dependencies only
# ca-certificates for TLS/SSL support
RUN apk add --no-cache ca-certificates

# Non-root user for security
RUN addgroup -g 3000 otel && \
    adduser -D -u 3000 -G otel otel && \
    chown -R otel:otel /app

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /otelcol /app/otelcol

# Copy OTel configuration
COPY --chown=otel:otel otel-collector-config.yml /app/config.yml

# Create directories for data and logs
RUN mkdir -p /app/data /app/logs && \
    chown -R otel:otel /app/data /app/logs

# Switch to non-root user
USER otel

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:13133/ || exit 1

# Expose OTLP gRPC port
EXPOSE 4317

# Expose OTLP HTTP port
EXPOSE 4318

# Expose metrics port (Prometheus exporter)
EXPOSE 8889

# Run OTel Collector
CMD ["/app/otelcol", "--config=/app/config.yml"]
