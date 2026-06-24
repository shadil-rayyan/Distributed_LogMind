# --- Stage 1: Build Binary ---
FROM golang:1.23-bookworm AS builder

# Install build essential tools required for CGO (SQLite compiling)
RUN apt-get update && apt-get install -y gcc g++ make libc6-dev --no-install-recommends && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy dependency manifests first for caching optimizations
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Build the Go app with optimizations and CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o logmind cmd/logmind/main.go

# --- Stage 2: Final Secure Runtime ---
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates --no-install-recommends && rm -rf /var/lib/apt/lists/*
RUN groupadd --system --gid 10001 logmind             && useradd --system --uid 10001 --gid 10001 --home-dir /app --create-home --shell /usr/sbin/nologin logmind             && mkdir -p /app/data             && chown -R logmind:logmind /app

WORKDIR /app

# Copy compiled binary and frontend files from builder stage
COPY --from=builder /app/logmind /app/logmind
COPY --from=builder /app/index.html /app/index.html
RUN chown logmind:logmind /app/logmind /app/index.html

# Memory limit via ENV for Go 1.19+ soft memory management
ENV GOMEMLIMIT=96MiB

# Expose production engine port
EXPOSE 8080

USER logmind:logmind

# Run the application
ENTRYPOINT ["/app/logmind"]
