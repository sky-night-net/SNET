# Step 1: Build the Go binary
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o skynet-v2 main.go

# Step 2: Final runtime image
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies: docker-cli to manage VPN containers, iproute2 for network management
RUN apk add --no-cache docker-cli iproute2 ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/skynet-v2 .

# Copy HTML templates and static assets
COPY --from=builder /app/web/html ./web/html

# Create folders for logs and database
RUN mkdir -p /etc/skynet /var/log/skynet

# Persistence volumes
VOLUME ["/etc/skynet", "/var/log/skynet"]

# Port for the panel and subscription server
EXPOSE 2053 2054

# Run SNET
CMD ["./skynet-v2", "run"]
