# --- STAGE 1: Build Stage ---
FROM golang:1.22-alpine AS builder

# Set working directory inside the container
WORKDIR /app

# Install git and certs (needed for downloading dependencies)
RUN apk add --no-cache git ca-certificates

# Copy go module files first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build a static, lightweight CGO-disabled Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bookstore-api ./cmd/api

# --- STAGE 2: Minimal Runtime Stage ---
FROM alpine:3.19

WORKDIR /app

# Install CA certificates for TLS/HTTPS outgoing connections
RUN apk --no-cache add ca-certificates

# Copy the compiled binary from the builder stage
COPY --from=builder /app/bookstore-api .

# Copy data files in case JSON storage mode is enabled locally
COPY --from=builder /app/data ./data

# Expose HTTP port
EXPOSE 8080

# Run the API binary
CMD ["./bookstore-api"]