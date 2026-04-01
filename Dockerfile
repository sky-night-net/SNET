# Step 1: Build the Go binary
FROM golang:alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev curl

# Create static directory and download assets
RUN mkdir -p web/static && \
    curl -L https://unpkg.com/vue@3.2.47/dist/vue.global.prod.js -o web/static/vue.js && \
    curl -L https://unpkg.com/ant-design-vue@3.2.20/dist/antd.min.js -o web/static/antd.js && \
    curl -L https://unpkg.com/ant-design-vue@3.2.20/dist/antd.min.css -o web/static/antd.css && \
    curl -L https://unpkg.com/axios@1.3.4/dist/axios.min.js -o web/static/axios.js && \
    curl -L https://unpkg.com/chart.js@3.9.1/dist/chart.min.js -o web/static/chart.js

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
COPY --from=builder /app/web/static ./web/static

# Create folders for logs and database
RUN mkdir -p /etc/snet /var/log/snet

# Volume definitions at the end
VOLUME ["/etc/snet", "/var/log/snet"]

# Run SNET
CMD ["./snet", "run"]
