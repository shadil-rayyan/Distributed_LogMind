# --- Stage 1: Build Binary ---
FROM golang:1.23-slim AS builder

# Install build essential tools required for CGO (SQLite compiling)
RUN apt-get update && apt-get install -y gcc g++ make libc6-dev --no-install-recommends && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy dependency manifests first for caching optimizations
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Build the Go app with optimizations and CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o logmind .

# --- Stage 2: Final Secure Runtime ---
FROM debian:bookworm-slim

WORKDIR /app

# Install runtime dependencies like CA certificates for external requests
RUN apt-get update && apt-get install -y ca-certificates --no-install-recommends && rm -rf /var/lib/apt/lists/*

# Copy compiled binary from builder stage
COPY --from=builder /app/logmind /app/logmind

# Expose production engine port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app/logmind"]