# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /message ./cmd/message

# Runtime stage
FROM alpine:3.21

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN adduser -D -u 1000 appuser
USER appuser

# Copy binary from builder
COPY --from=builder /message /app/message

ENTRYPOINT ["/app/message"]
