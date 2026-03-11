# Build stage
FROM oven/bun:1 AS builder

WORKDIR /app

# Build arguments for NEXT_PUBLIC_ variables (baked in at build time)
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL

# Copy package files
COPY package.json bun.lockb ./

# Install dependencies
RUN bun install --frozen-lockfile

# Copy source code
COPY . .

# Build the application
ENV NEXT_TELEMETRY_DISABLED=1
# Server env vars (WorkOS, etc.) are injected at runtime, not build time
ENV SKIP_ENV_VALIDATION=1
RUN bun run build

# Runtime stage
FROM oven/bun:1-slim

WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV HOSTNAME=0.0.0.0

# Copy necessary files from builder
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

CMD ["bun", "run", "server.js"]
