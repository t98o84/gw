# Development Dockerfile
FROM golang:1.23-alpine

# Install dependencies
RUN apk add --no-cache git fzf

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum* ./
RUN go mod download || true

# Copy source code
COPY . .

# Default command
CMD ["go", "build", "-o", "gw", "."]
