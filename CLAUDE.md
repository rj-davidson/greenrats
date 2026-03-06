# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Commands

### Full Stack Development
```bash
# Start entire stack (from repo root)
docker compose up --build

# View: http://localhost:3000 (frontend) and http://localhost:8000 (backend)
```

### Backend (Go with Ent & Fiber)
```bash
# From /backend directory

# Development
go mod download                            # Install dependencies
air                                        # Start API with hot reload
air -c .air.ingest.toml                    # Start ingest service with hot reload

# Build
go build -o bin/api ./cmd/api              # Build API binary
go build -o bin/ingest ./cmd/ingest        # Build ingest binary

# Run directly
go run ./cmd/api                           # Run API server
go run ./cmd/ingest                        # Run ingest service

# Testing
go test ./...                              # Run all tests
go test ./internal/features/picks/...      # Test specific feature
go test -run TestPickValidation ./...      # Run tests matching pattern
go test -v ./...                           # Verbose output
go test -race ./...                        # Run with race detector
go test -cover ./...                       # Run with coverage

# Linting & Formatting
golangci-lint run                          # Run all linters
golangci-lint run --fix                    # Autofix where possible

# Ent (ORM)
go generate ./ent                          # Regenerate Ent code after schema changes

# Atlas (Migrations)
atlas migrate diff <name> --env local      # Generate new migration after schema changes
atlas migrate apply --env local            # Apply pending migrations
atlas migrate status --env local           # Check migration status
atlas migrate hash --env local             # Rehash migration directory (after manual edits)

# Dependencies
go get <package>                           # Add dependency
go get <package>@latest                    # Update dependency
go mod tidy                                # Clean up go.mod and go.sum
```

### Frontend (Next.js/React)
```bash
# From /frontend directory
bun install                                # Install dependencies
bun dev                                    # Start dev server
bun run build                              # Production build
bun run typegen                            # Generate Next.js types
bun run tsc                                # Type check without emitting

# Linting & Formatting
bun run lint                               # oxlint + eslint with fixes
bun run format                             # Format with Prettier
bun run format:check                       # Check formatting

# Storybook
bun run storybook                          # Start Storybook on :6006
bun run build-storybook                    # Build Storybook

# Dependencies
bun add <package-name>                     # Add dependency
bun remove <package-name>                  # Remove dependency
```

## Code Style

**Comments Policy**:
- All code should be idiomatic and self-documenting
- Do NOT add code comments unless the code is non-idiomatic (which should be avoided)
- If code requires a comment to be understood, refactor it to be clearer instead
- The only acceptable comments are for truly unavoidable non-idiomatic code (e.g., workarounds for external bugs, performance optimizations that sacrifice clarity)

## Product Safety (No Wagering + Brand Safety)

- No entry fees, prize pools, payouts, escrow, or pot tracking anywhere in the product
- Avoid gambling language (bet, wager, odds, parlay, payout) in UI/UX and copy
- Do not use PGA TOUR logos or marks; use descriptive references and avoid any implied endorsement
- If a feature might touch prizes or payments between users, require legal review before build

## Architecture Overview

### Monorepo Structure
- **Root**: docker-compose.yml orchestrates the full stack
- **Backend** (`/backend`): Go services with Fiber (HTTP) and Ent (ORM)
- **Frontend** (`/frontend`): Next.js 16 App Router with React 19

### Backend Architecture

**Language**: Go 1.25

**Key Frameworks & Libraries**:
- **HTTP**: Fiber v2
- **ORM**: Ent with Atlas versioned migrations
- **Config**: Viper (environment-based configuration)
- **HTTP Client**: Resty (for external API calls)
- **Authentication**: WorkOS JWT validation

**Two Executables**:
- **`cmd/api`**: HTTP API server (Fiber), handles user requests, authentication, SSE streams
- **`cmd/ingest`**: Data collection service, runs scheduled goroutines during tournaments to fetch external data

**Feature Module Pattern** (`internal/features/<name>/`):
Each feature is self-contained:
- `handlers.go` - Fiber route handlers (keep thin, delegate to services)
- `service.go` - Business logic
- `types.go` - Request/response structs, validation
- `*_test.go` - Feature tests

Features: admin, fields, golfers, leagues, leaderboards, picks, tournaments, users

**External API Clients** (`internal/external/`):
- `balldontlie/` - Golf statistics and tournament data (BallDontLie.io)
- `googlemaps/` - Google Maps integration
- `pgatour/` - PGA Tour data

**Key Conventions**:
- Handlers are thin wrappers—push logic to services
- Ent models are shared across both executables
- Use dependency injection (pass services to handlers)
- Errors should be wrapped with context: `fmt.Errorf("failed to get pick: %w", err)`

**Logging**:
- Use Go's built-in `log/slog` package exclusively (no standard `log` package)
- All services and clients receive a `*slog.Logger` via constructor injection
- Log levels: Info (business operations), Debug (dev details, sparingly), Warn (degraded states), Error (leave to callers)
- Development mode uses `slog.LevelDebug`, production uses `slog.LevelInfo`
- For tests, use a discard logger: `slog.New(slog.NewTextHandler(io.Discard, nil))`
- Don't log: expected client failures, per-iteration progress, skipped operations, redundant context

**Data Model Notes**:
- Tournament and golfer data is **shared** across all leagues
- Picks and leaderboards are **league-scoped**
- Users can belong to multiple leagues
- One pick per user per tournament; golfers cannot be reused within a league-year

### Frontend Architecture

**Framework**: Next.js 16 with App Router, React 19

**Key Infrastructure**:
- **Authentication**: WorkOS AuthKit (`@workos-inc/authkit-nextjs`)
- **State Management**: TanStack Query v5 for server state, TanStack Store for local state
- **Styling**: Tailwind CSS v4 with tailwindcss-animate
- **UI Components**: Radix UI primitives + shadcn/ui generated components
- **Forms**: React Hook Form with Zod validation
- **Real-time**: SSE (Server-Sent Events) for live tournament updates

**Feature Module Pattern** (`features/<name>/`):
- `queries.ts` - TanStack Query hooks/mutations
- `types.ts` - Shared types/Zod schemas
- `components/` - Feature-specific UI components

Features: admin, dashboard, golfers, leaderboards, leagues, payments, picks, tournaments, users

**Key Conventions**:
- **DO NOT hand-edit** `components/shadcn/` - these are generated; reconfigure Shadcn instead
- Prefer feature-local components over promoting to `components/core/`
- Use TanStack Query for all server state (see `lib/query/`)
- Use SSE hooks from `lib/sse/` for live tournament data
- oxlint (fast) + eslint (thorough) via `bun run lint`
- Prettier for formatting (2-space indent)
- Components in PascalCase, hooks/utils in camelCase

### Communication Between Frontend & Backend

- **HTTP API**: Frontend uses `lib/query/api-client.ts` to call Fiber REST endpoints
- **SSE**: Live tournament updates streamed from backend SSE endpoints
- **Authentication**: WorkOS JWT tokens passed in Authorization header

## Development Workflow

1. **Setup**: Clone, install dependencies, populate `.env` files from `.env.example` / `.env.local.example`
2. **Running Locally**: `docker compose up --build` from repo root
3. **Backend Changes**:
   - Make code changes in feature modules
   - If Ent schemas changed: `go generate ./ent` → `atlas migrate diff <name> --env local` → review SQL → `atlas migrate apply --env local`
   - Run tests: `go test ./...`
   - Lint: `golangci-lint run`
4. **Frontend Changes**:
   - Make code changes in features or components
   - Type check: `bun run tsc`
   - Lint/format: `bun run lint && bun run format`
5. **Commits**: Use short, imperative style (e.g., "Add pick validation logic")

## Testing

### Backend (Go)
- Use standard library `testing` package
- Use `testify/assert` and `testify/require` for assertions
- Use `testify/mock` for mocking interfaces
- Table-driven tests preferred for multiple cases
- Tests live alongside code (`*_test.go` files)

### Frontend
- Vitest for unit/interaction tests
- Storybook for visual component testing

## Environment & Configuration

- Backend: `/backend/.env` (see `.env.example`, loaded via Viper)
- Frontend: `/frontend/.env.local` (see `.env.local.example`)
- Never commit secrets—populate from secure source

## Deployment

- **Platform**: Railway.app for CI/CD and hosting
- **Services**: API, Ingest, Frontend, PostgreSQL
- **Environments**: Development (local), Staging, Production
