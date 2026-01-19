# Build stage
FROM golang:1.25-alpine AS builder

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

# Install runtime dependencies and Atlas CLI
RUN apk add --no-cache ca-certificates tzdata curl \
    && curl -sSf https://atlasgo.sh | sh

# Copy binaries from builder
COPY --from=builder /bin/api /app/api
COPY --from=builder /bin/ingest /app/ingest

# Copy migrations and Atlas config
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/atlas.hcl /app/atlas.hcl

# Create entrypoint script that runs migrations then starts API
RUN printf '#!/bin/sh\nset -e\natlas migrate apply --env production --url "$DATABASE_URL"\nexec /app/api\n' > /app/entrypoint.sh \
    && chmod +x /app/entrypoint.sh

CMD ["/app/entrypoint.sh"]
