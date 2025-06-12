# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments for version info
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -o mcp-mindport .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 mindport && \
    adduser -D -s /bin/sh -u 1001 -G mindport mindport

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/mcp-mindport .

# Create data directories with proper permissions
RUN mkdir -p /app/data/storage /app/data/search && \
    chown -R mindport:mindport /app

# Switch to non-root user
USER mindport

# Expose default port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ./mcp-mindport --health-check || exit 1

# Set entrypoint
ENTRYPOINT ["./mcp-mindport"]

# Default command
CMD ["--help"]