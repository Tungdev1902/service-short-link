# Development Dockerfile
FROM golang:1.24-alpine

# Install development dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    curl \
    bash \
    make \
    vim \
    nano \
    htop \
    tree \
    jq \
    mariadb-client \
    redis \
    inotify-tools \
    openssl

# Set working directory
WORKDIR /app

# Install Go tools for development
RUN go install github.com/air-verse/air@v1.62.0 && \
    go install github.com/swaggo/swag/cmd/swag@latest && \
    go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod tidy

# Create vendor directory
RUN go mod vendor

COPY .air.toml .air.toml


# Expose port
EXPOSE 8080

# Default command - run the built binary
WORKDIR /app
CMD ["sh", "-c", "./shortlink"]
