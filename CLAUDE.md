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
go fmt ./...                               # Format code (also run by linter)
goimports -w .                             # Organize imports

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

**External Data Sources**:
- **Scratch Golf API**: Tournament and golfer data
- **BallDontLie.io**: Additional golf statistics

**Directory Layout**:
```
backend/
  cmd/
    api/
      main.go                  # API server entrypoint
    ingest/
      main.go                  # Data ingestion service entrypoint
  internal/
    config/                    # Viper configuration loading
    server/                    # Fiber app setup, middleware, routes
    auth/                      # WorkOS JWT authentication middleware
    sse/                       # Server-Sent Events for live updates
    features/
      tournaments/             # Tournament schedules, data
        handlers.go            # Fiber route handlers
        service.go             # Business logic
        types.go               # Request/response types
      golfers/                 # Golfer profiles, stats
      picks/                   # User pick logic (one per tournament, no repeats)
      leagues/                 # League management
      leaderboards/            # Earnings calculations, rankings
      users/                   # User profiles
    external/                  # External API clients
      scratchgolf/             # Scratch Golf API client
      balldontlie/             # BallDontLie.io client
  ent/
    schema/                    # Ent schema definitions
    ...                        # Generated Ent code (do not edit)
  migrations/                  # Atlas versioned migration files
  atlas.hcl                    # Atlas configuration
  .air.toml                    # Air config for API hot reload
  .air.ingest.toml             # Air config for ingest hot reload
```

**Two Executables**:
- **`cmd/api`**: HTTP API server (Fiber), handles user requests, authentication, SSE streams
- **`cmd/ingest`**: Data collection service, runs scheduled goroutines during tournaments to fetch external data

**Feature Module Pattern**:
Each feature is self-contained:
- `handlers.go` - Fiber route handlers (keep thin, delegate to services)
- `service.go` - Business logic
- `types.go` - Request/response structs, validation
- `*_test.go` - Feature tests

**Key Conventions**:
- Handlers are thin wrappers—push logic to services
- Ent models are shared across both executables
- External API clients live in `internal/external/`
- Use dependency injection (pass services to handlers)
- Errors should be wrapped with context: `fmt.Errorf("failed to get pick: %w", err)`

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

**Directory Layout**:
```
frontend/
  app/                         # Next.js App Router
    (dashboard)/               # Route group for authenticated pages
    login/, callback/          # Auth flows
    layout.tsx, globals.css    # Root layout & global styles
  features/                    # Feature-organized code
    tournaments/
      queries.ts               # TanStack Query hooks/mutations
      types.ts                 # Shared types/Zod schemas
      components/              # Feature-specific UI components
    picks/
    leagues/
    leaderboards/
    users/
  components/
    shadcn/                    # Generated Shadcn UI (DO NOT hand-edit)
    core/                      # Shared app-wide components
  lib/
    query/                     # API client, requestor, query setup
    sse/                       # SSE client utilities for live updates
    providers/                 # App-level providers
    hooks/                     # Reusable hooks
    types/                     # Shared types
  stories/                     # Storybook stories
  .storybook/                  # Storybook configuration
```

**Feature Module Pattern**:
- Data fetching/mutations in `queries.ts` at feature root
- Types/schemas in `types.ts`
- Feature-specific UI in `components/` subdirectory
- Keep `components/core/` for truly shared, app-wide components only

**Key Conventions**:
- **DO NOT hand-edit** `components/shadcn/` - these are generated; reconfigure Shadcn instead
- Prefer feature-local components over promoting to `components/core/`
- Use TanStack Query for all server state (see `lib/query/`)
- Use SSE hooks from `lib/sse/` for live tournament data

**Linting/Formatting**:
- oxlint (fast) + eslint (thorough) via `bun run lint`
- Prettier for formatting (2-space indent)
- Components in PascalCase, hooks/utils in camelCase

### Communication Between Frontend & Backend

- **HTTP API**: Frontend uses `lib/query/api-client.ts` to call Fiber REST endpoints
- **SSE**: Live tournament updates streamed from `/api/tournaments/:id/live` endpoint
- **Authentication**: WorkOS JWT tokens passed in Authorization header

### Real-time Updates (SSE)

The app uses Server-Sent Events for live tournament data during active tournaments:
- Backend: `internal/sse/` manages SSE connections and broadcasts
- Frontend: `lib/sse/` provides hooks for subscribing to tournament updates
- TanStack Query integrates with SSE to automatically update cached data

## Development Workflow

1. **Setup**: Follow root README.md - clone, install dependencies, populate `.env` files
2. **Running Locally**: `docker compose up --build` from repo root
3. **Backend Changes**:
   - Make code changes in feature modules
   - If Ent schemas changed:
     - Run `go generate ./ent` to regenerate Ent code
     - Run `atlas migrate diff <name> --env local` to create migration
     - Review generated SQL in `migrations/`
     - Run `atlas migrate apply --env local`
   - Run tests: `go test ./...`
   - Lint: `golangci-lint run`
4. **Frontend Changes**:
   - Make code changes in features or components
   - Type check: `bun run tsc`
   - Lint/format: `bun run lint && bun run format`
5. **Commits**: Use short, imperative style (e.g., "Add pick validation logic")

## Product Context

**Purpose**: GreenRats is a golf pick'em website where users pick one golfer per PGA Tour/major tournament throughout the season.

**Core Mechanics**:
- Users pick one golfer per tournament
- Once a golfer is picked, they cannot be reused for the rest of the season
- Leaderboards track cumulative earnings across all tournaments
- Users compete within leagues

**Key Entities**:
- **Tournaments**: PGA Tour events and majors (shared across all leagues)
- **Golfers**: Player profiles and stats (shared across all leagues)
- **Picks**: User selections (league-scoped, one per tournament, no repeats)
- **Leagues**: User groups with shared leaderboards
- **Leaderboards**: Cumulative earnings rankings per league

## Testing

### Backend (Go)
- Use standard library `testing` package
- Use `testify/assert` and `testify/require` for assertions
- Use `testify/mock` for mocking interfaces
- Table-driven tests preferred for multiple cases
- Tests live alongside code (`*_test.go` files)

```go
func TestPickService_ValidatePick(t *testing.T) {
    tests := []struct {
        name    string
        pick    Pick
        wantErr bool
    }{
        {"valid pick", Pick{...}, false},
        {"duplicate golfer", Pick{...}, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

### Frontend
- Vitest for unit/interaction tests
- Storybook for visual component testing

## Linting Configuration

### Backend (golangci-lint)
Recommended `.golangci.yml` configuration:
```yaml
run:
  timeout: 5m
  go: "1.25"

linters:
  enable:
    - errcheck        # Check for unchecked errors
    - gosimple        # Simplify code
    - govet           # Suspicious constructs
    - ineffassign     # Unused assignments
    - staticcheck     # Static analysis
    - unused          # Unused code
    - gofmt           # Formatting
    - goimports       # Import organization
    - misspell        # Spelling mistakes
    - gocritic        # Opinionated linter
    - errname         # Error naming conventions
    - errorlint       # Error wrapping
    - wrapcheck       # Error wrapping in public funcs

linters-settings:
  errcheck:
    check-type-assertions: true
  gocritic:
    enabled-tags:
      - diagnostic
      - style
      - performance
```

### Pre-commit Hooks
Recommended `.pre-commit-config.yaml`:
```yaml
repos:
  - repo: local
    hooks:
      - id: go-fmt
        name: go fmt
        entry: go fmt ./...
        language: system
        files: \.go$
        pass_filenames: false
      - id: go-vet
        name: go vet
        entry: go vet ./...
        language: system
        files: \.go$
        pass_filenames: false
      - id: golangci-lint
        name: golangci-lint
        entry: golangci-lint run
        language: system
        files: \.go$
        pass_filenames: false
      - id: go-test
        name: go test
        entry: go test -short ./...
        language: system
        files: \.go$
        pass_filenames: false
```

## Environment & Configuration

- Backend: `/backend/.env` (loaded via Viper)
- Frontend: `/frontend/.env.local`
- Never commit secrets—populate from secure source

**Backend Config Pattern (Viper)**:
```go
type Config struct {
    Port        int    `mapstructure:"PORT"`
    DatabaseURL string `mapstructure:"DATABASE_URL"`
    WorkOSClientID string `mapstructure:"WORKOS_CLIENT_ID"`
    // ...
}
```

## Deployment

- **Platform**: Railway.app for CI/CD and hosting
- **Services**: API, Ingest (scheduled during tournaments), PostgreSQL
- **Environments**: Development (local), Staging, Production

## Atlas Migration Workflow

```bash
# After modifying ent/schema/*.go files:

# 1. Regenerate Ent code
go generate ./ent

# 2. Generate migration (compares schema to migration directory)
atlas migrate diff add_leagues_table --env local

# 3. Review the generated SQL file in migrations/
# Look for any destructive operations (DROP, ALTER removing columns)

# 4. Apply migration
atlas migrate apply --env local

# 5. If you need to edit a migration manually, rehash afterward
atlas migrate hash --env local
```

**atlas.hcl example**:
```hcl
env "local" {
  src = "ent://ent/schema"
  dev = "docker://postgres/16/dev?search_path=public"
  migration {
    dir = "file://migrations"
  }
  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
```
