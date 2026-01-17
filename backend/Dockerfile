# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the API binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api

# Build the ingest binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/ingest ./cmd/ingest

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binaries from builder
COPY --from=builder /bin/api /app/api
COPY --from=builder /bin/ingest /app/ingest

# Copy migrations if needed
COPY --from=builder /app/migrations /app/migrations

# Default to running the API
CMD ["/app/api"]
