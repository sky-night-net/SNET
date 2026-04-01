# Step 1: Build the Go binary
FROM golang:alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go.mod and optional go.sum, then download dependencies
COPY go.mod go.sum* ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Reconstruct go.sum and tidy modules since we renamed the git repo
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o snet main.go

# Step 2: Final runtime image
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies: docker-cli to manage VPN containers, iproute2 for network management
RUN apk add --no-cache docker-cli iproute2 ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/snet .

# Copy HTML templates and static assets
COPY --from=builder /app/web/html ./web/html

# Create folders for logs and database
RUN mkdir -p /etc/snet /var/log/snet

# Persistence volumes
VOLUME ["/etc/snet", "/var/log/snet"]

# Port for the panel and subscription server
EXPOSE 2053 2054

# Run SNET
CMD ["./snet", "run"]
