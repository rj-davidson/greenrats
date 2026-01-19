# GreenRats Repository Setup Checklist

This document provides step-by-step instructions for Claude Code to set up the greenrats repository from scratch. Each step should be committed before moving to the next.

---

## Step 1: Initialize Root Repository Structure

**Goal**: Create the basic monorepo structure with root configuration files.

- [ ] Create root `.gitignore` with common ignores for Go, Node.js, and IDE files
- [ ] Create root `README.md` with project overview and setup instructions
- [ ] Create empty `/backend` directory with `.gitkeep`
- [ ] Create empty `/frontend` directory with `.gitkeep`
- [ ] Create `docker-compose.yml` with PostgreSQL service (other services will be added later)

```bash
git add .
git commit -m "Initialize monorepo structure"
```

---

## Step 2: Initialize Go Backend Module

**Goal**: Set up the Go module with basic directory structure.

- [ ] Initialize Go module: `go mod init github.com/rj-davidson/greenrats`
- [ ] Create directory structure:
  ```
  backend/
    cmd/
      api/.gitkeep
      ingest/.gitkeep
    internal/
      config/.gitkeep
      server/.gitkeep
      auth/.gitkeep
      sse/.gitkeep
      features/
        tournaments/.gitkeep
        golfers/.gitkeep
        picks/.gitkeep
        leagues/.gitkeep
        leaderboards/.gitkeep
        users/.gitkeep
      external/
        livegolfdata/.gitkeep
        balldontlie/.gitkeep
    ent/.gitkeep
    migrations/.gitkeep
  ```
- [ ] Create `/backend/.env.example` with placeholder environment variables:
  ```
  PORT=8000
  DATABASE_URL=postgres://postgres:postgres@localhost:5432/greenrats?sslmode=disable
  WORKOS_CLIENT_ID=
  WORKOS_API_KEY=
  LIVE_GOLF_DATA_API_KEY=
  LIVE_GOLF_DATA_BASE_URL=https://live-golf-data.p.rapidapi.com
  BALL_DONT_LIE_API_KEY=
  ```
- [ ] Create `/backend/.gitignore` for Go-specific ignores (bin/, .env, etc.)

```bash
git add .
git commit -m "Initialize Go backend module structure"
```

---

## Step 3: Add Core Go Dependencies

**Goal**: Add all required Go dependencies to go.mod.

- [ ] Add Fiber v2: `go get github.com/gofiber/fiber/v2`
- [ ] Add Ent: `go get entgo.io/ent/cmd/ent`
- [ ] Add Atlas: `go get ariga.io/atlas`
- [ ] Add Viper: `go get github.com/spf13/viper`
- [ ] Add Resty: `go get github.com/go-resty/resty/v2`
- [ ] Add testify: `go get github.com/stretchr/testify`
- [ ] Add WorkOS Go SDK: `go get github.com/workos/workos-go/v4`
- [ ] Run `go mod tidy`

```bash
git add .
git commit -m "Add core Go dependencies"
```

---

## Step 4: Set Up Viper Configuration

**Goal**: Create configuration loading with Viper.

- [ ] Create `/backend/internal/config/config.go`:
  - Define `Config` struct with all environment variables
  - Create `Load()` function that reads from environment/.env file
  - Include fields: Port, DatabaseURL, WorkOS credentials, external API keys
- [ ] Create `/backend/internal/config/config_test.go` with basic tests

```bash
git add .
git commit -m "Add Viper configuration loading"
```

---

## Step 5: Initialize Ent Schemas

**Goal**: Create initial Ent schemas for core entities.

- [ ] Initialize Ent: `go run -mod=mod entgo.io/ent/cmd/ent new User Tournament Golfer Pick League LeagueMemebership`
- [ ] Edit `/backend/ent/schema/user.go`:
  - Fields: id (uuid), workos_id (string, unique), email (string), display_name (string), created_at, updated_at
  - Edges: picks, league_memberships
- [ ] Edit `/backend/ent/schema/tournament.go`:
  - Fields: id (uuid), external_id (string, unique), name (string), start_date (time), end_date (time), status (enum: upcoming/active/completed), season_year (int), created_at, updated_at
  - Edges: picks, golfer_tournaments
- [ ] Edit `/backend/ent/schema/golfer.go`:
  - Fields: id (uuid), external_id (string, unique), name (string), country (string), world_ranking (int), image_url (string, optional), created_at, updated_at
  - Edges: picks, golfer_tournaments
- [ ] Edit `/backend/ent/schema/pick.go`:
  - Fields: id (uuid), created_at
  - Edges: user, tournament, golfer, league
  - Indexes: unique on (user_id, tournament_id, league_id), unique on (user_id, golfer_id, league_id, season_year)
- [ ] Edit `/backend/ent/schema/league.go`:
  - Fields: id (uuid), name (string), code (string, unique), season_year (int), created_at, updated_at
  - Edges: memberships, picks
- [ ] Edit `/backend/ent/schema/leaguemembership.go`:
  - Fields: id (uuid), role (enum: owner/member), joined_at
  - Edges: user, league
  - Indexes: unique on (user_id, league_id)
- [ ] Generate Ent code: `go generate ./ent`

```bash
git add .
git commit -m "Add initial Ent schemas"
```

---

## Step 6: Set Up Atlas Migrations

**Goal**: Configure Atlas for versioned migrations.

- [ ] Create `/backend/atlas.hcl`:
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

  env "production" {
    src = "ent://ent/schema"
    url = getenv("DATABASE_URL")
    migration {
      dir = "file://migrations"
    }
  }
  ```
- [ ] Generate initial migration: `atlas migrate diff initial_schema --env local`
- [ ] Review generated migration file in `/backend/migrations/`

```bash
git add .
git commit -m "Configure Atlas migrations with initial schema"
```

---

## Step 7: Create Fiber Server Setup

**Goal**: Set up the basic Fiber HTTP server with middleware.

- [ ] Create `/backend/internal/server/server.go`:
  - Create `Server` struct holding Fiber app, config, and Ent client
  - Create `New()` constructor
  - Create `Start()` method to run the server
  - Create `Shutdown()` method for graceful shutdown
- [ ] Create `/backend/internal/server/middleware.go`:
  - Add CORS middleware configuration
  - Add request logging middleware
  - Add recover middleware
- [ ] Create `/backend/internal/server/routes.go`:
  - Set up route groups: `/api/v1/`
  - Add health check endpoint: `GET /health`
  - Placeholder route registration for features

```bash
git add .
git commit -m "Add Fiber server setup with middleware"
```

---

## Step 8: Create API Entrypoint

**Goal**: Create the main.go for the API service.

- [ ] Create `/backend/cmd/api/main.go`:
  - Load configuration with Viper
  - Initialize Ent client with database connection
  - Run Atlas migrations on startup (optional, controlled by env var)
  - Initialize and start Fiber server
  - Handle graceful shutdown on SIGINT/SIGTERM
- [ ] Verify it compiles: `go build ./cmd/api`

```bash
git add .
git commit -m "Add API service entrypoint"
```

---

## Step 9: Create Ingest Service Entrypoint

**Goal**: Create the main.go for the ingest service with goroutine scheduling.

- [ ] Create `/backend/cmd/ingest/main.go`:
  - Load configuration with Viper
  - Initialize Ent client
  - Set up scheduled goroutines for data collection
  - Handle graceful shutdown
- [ ] Create `/backend/internal/external/livegolfdata/client.go`:
  - Create Resty client wrapper
  - Define placeholder methods for API calls
- [ ] Create `/backend/internal/external/balldontlie/client.go`:
  - Create Resty client wrapper
  - Define placeholder methods for API calls

```bash
git add .
git commit -m "Add ingest service entrypoint with external API clients"
```

---

## Step 10: Add SSE Support

**Goal**: Set up Server-Sent Events infrastructure.

- [ ] Create `/backend/internal/sse/broker.go`:
  - Create `Broker` struct to manage SSE connections
  - Implement `Subscribe()` method for new clients
  - Implement `Unsubscribe()` method for disconnects
  - Implement `Broadcast()` method to send events to all clients
- [ ] Create `/backend/internal/sse/handler.go`:
  - Create Fiber handler for SSE endpoint
  - Handle connection lifecycle
- [ ] Register SSE routes in server routes

```bash
git add .
git commit -m "Add SSE infrastructure for live updates"
```

---

## Step 11: Add WorkOS Authentication Middleware

**Goal**: Set up JWT validation for protected routes.

- [ ] Create `/backend/internal/auth/workos.go`:
  - Create middleware function to validate WorkOS JWT
  - Extract user ID from token and add to Fiber context
  - Handle token refresh if needed
- [ ] Create `/backend/internal/auth/context.go`:
  - Helper functions to get user from context
- [ ] Update routes to use auth middleware on protected endpoints

```bash
git add .
git commit -m "Add WorkOS JWT authentication middleware"
```

---

## Step 12: Add Air Hot Reload Configuration

**Goal**: Configure Air for development hot reloading.

- [ ] Create `/backend/.air.toml` for API service:
  ```toml
  root = "."
  tmp_dir = "tmp"

  [build]
  cmd = "go build -o ./tmp/api ./cmd/api"
  bin = "./tmp/api"
  full_bin = "./tmp/api"
  include_ext = ["go", "tpl", "tmpl", "html"]
  exclude_dir = ["assets", "tmp", "vendor", "frontend", "node_modules"]
  exclude_regex = ["_test.go"]
  delay = 1000

  [log]
  time = false

  [color]
  main = "magenta"
  watcher = "cyan"
  build = "yellow"
  runner = "green"

  [misc]
  clean_on_exit = true
  ```
- [ ] Create `/backend/.air.ingest.toml` for ingest service (similar, but builds ingest binary)

```bash
git add .
git commit -m "Add Air hot reload configuration"
```

---

## Step 13: Add Go Linting Configuration

**Goal**: Set up golangci-lint with recommended settings.

- [ ] Create `/backend/.golangci.yml`:
  ```yaml
  run:
    timeout: 5m
    go: "1.23"

  linters:
    enable:
      - errcheck
      - gosimple
      - govet
      - ineffassign
      - staticcheck
      - unused
      - gofmt
      - goimports
      - misspell
      - gocritic
      - errname
      - errorlint

  linters-settings:
    errcheck:
      check-type-assertions: true
    gocritic:
      enabled-tags:
        - diagnostic
        - style
        - performance

  issues:
    exclude-dirs:
      - ent
  ```
- [ ] Verify linting works: `golangci-lint run`

```bash
git add .
git commit -m "Add golangci-lint configuration"
```

---

## Step 14: Add Pre-commit Hooks

**Goal**: Set up pre-commit hooks for code quality.

- [ ] Create `/backend/.pre-commit-config.yaml`:
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
  ```

```bash
git add .
git commit -m "Add pre-commit hooks configuration"
```

---

## Step 15: Initialize Next.js Frontend

**Goal**: Create the Next.js application with base configuration.

- [ ] From `/frontend` directory, run: `bunx create-next-app@latest . --typescript --tailwind --eslint --app --src-dir=false --import-alias="@/*"`
- [ ] Remove default boilerplate content from `app/page.tsx`
- [ ] Update `app/layout.tsx` with proper metadata for GreenRats
- [ ] Verify it runs: `bun dev`

```bash
git add .
git commit -m "Initialize Next.js frontend"
```

---

## Step 16: Add Frontend Dependencies

**Goal**: Add all required frontend dependencies.

- [ ] Add TanStack Query: `bun add @tanstack/react-query`
- [ ] Add TanStack Store: `bun add @tanstack/store @tanstack/react-store`
- [ ] Add React Hook Form + Zod: `bun add react-hook-form @hookform/resolvers zod`
- [ ] Add WorkOS AuthKit: `bun add @workos-inc/authkit-nextjs`
- [ ] Add Radix UI primitives: `bun add @radix-ui/react-slot @radix-ui/react-dialog @radix-ui/react-dropdown-menu @radix-ui/react-avatar @radix-ui/react-select @radix-ui/react-tabs`
- [ ] Add utility libraries: `bun add clsx tailwind-merge lucide-react`
- [ ] Add dev dependencies: `bun add -d @types/node oxlint prettier prettier-plugin-tailwindcss`

```bash
git add .
git commit -m "Add frontend dependencies"
```

---

## Step 17: Set Up Shadcn UI

**Goal**: Initialize shadcn/ui with base components.

- [ ] Initialize shadcn: `bunx shadcn@latest init`
  - Select default style, slate color, CSS variables: yes
  - Configure to output to `components/shadcn/`
- [ ] Add base components:
  ```bash
  bunx shadcn@latest add button
  bunx shadcn@latest add card
  bunx shadcn@latest add input
  bunx shadcn@latest add label
  bunx shadcn@latest add dialog
  bunx shadcn@latest add dropdown-menu
  bunx shadcn@latest add avatar
  bunx shadcn@latest add select
  bunx shadcn@latest add tabs
  bunx shadcn@latest add table
  bunx shadcn@latest add badge
  bunx shadcn@latest add skeleton
  bunx shadcn@latest add toast
  ```
- [ ] Create `/frontend/lib/utils.ts` with `cn()` helper if not created

```bash
git add .
git commit -m "Initialize shadcn/ui with base components"
```

---

## Step 18: Create Frontend Directory Structure

**Goal**: Set up feature-based directory structure.

- [ ] Create directory structure:
  ```
  frontend/
    app/
      (dashboard)/
        layout.tsx
        page.tsx
        tournaments/
          page.tsx
          [id]/
            page.tsx
        picks/
          page.tsx
        leagues/
          page.tsx
          [id]/
            page.tsx
        leaderboards/
          page.tsx
        settings/
          page.tsx
      login/
        page.tsx
      callback/
        page.tsx
      api/
        auth/
          [...all]/
            route.ts
    features/
      tournaments/
        queries.ts
        types.ts
        components/.gitkeep
      picks/
        queries.ts
        types.ts
        components/.gitkeep
      leagues/
        queries.ts
        types.ts
        components/.gitkeep
      leaderboards/
        queries.ts
        types.ts
        components/.gitkeep
      users/
        queries.ts
        types.ts
        components/.gitkeep
    components/
      core/.gitkeep
    lib/
      query/
        api-client.ts
        query-client.ts
      sse/
        use-sse.ts
      providers/
        query-provider.tsx
      hooks/.gitkeep
      types/.gitkeep
  ```
- [ ] Create placeholder files with basic exports

```bash
git add .
git commit -m "Create frontend feature directory structure"
```

---

## Step 19: Set Up TanStack Query

**Goal**: Configure TanStack Query with API client.

- [ ] Create `/frontend/lib/query/api-client.ts`:
  - Create fetch wrapper with base URL configuration
  - Add auth header injection
  - Add error handling
- [ ] Create `/frontend/lib/query/query-client.ts`:
  - Create and export QueryClient with default options
- [ ] Create `/frontend/lib/providers/query-provider.tsx`:
  - Create QueryClientProvider wrapper component
- [ ] Update `/frontend/app/layout.tsx` to include QueryProvider

```bash
git add .
git commit -m "Set up TanStack Query with API client"
```

---

## Step 20: Set Up SSE Client

**Goal**: Create SSE utilities for live tournament updates.

- [ ] Create `/frontend/lib/sse/use-sse.ts`:
  - Create `useSSE` hook that connects to SSE endpoint
  - Handle connection lifecycle (connect, reconnect, cleanup)
  - Parse incoming events and update TanStack Query cache
- [ ] Create `/frontend/lib/sse/types.ts`:
  - Define SSE event types

```bash
git add .
git commit -m "Add SSE client utilities"
```

---

## Step 21: Set Up WorkOS Authentication (Frontend)

**Goal**: Configure WorkOS AuthKit for authentication.

- [ ] Create `/frontend/.env.local.example`:
  ```
  NEXT_PUBLIC_API_URL=http://localhost:8000
  WORKOS_CLIENT_ID=
  WORKOS_API_KEY=
  WORKOS_COOKIE_PASSWORD=
  NEXT_PUBLIC_WORKOS_REDIRECT_URI=http://localhost:3000/callback
  ```
- [ ] Create `/frontend/app/api/auth/[...all]/route.ts`:
  - Set up WorkOS AuthKit route handler
- [ ] Create `/frontend/app/login/page.tsx`:
  - Create login page with WorkOS redirect
- [ ] Create `/frontend/app/callback/page.tsx`:
  - Handle OAuth callback
- [ ] Create `/frontend/lib/auth.ts`:
  - Export auth helpers (getUser, signOut, etc.)

```bash
git add .
git commit -m "Set up WorkOS authentication"
```

---

## Step 22: Configure Frontend Linting

**Goal**: Set up oxlint, ESLint, and Prettier.

- [ ] Create `/frontend/.prettierrc`:
  ```json
  {
    "semi": true,
    "singleQuote": false,
    "tabWidth": 2,
    "trailingComma": "es5",
    "plugins": ["prettier-plugin-tailwindcss"]
  }
  ```
- [ ] Create `/frontend/.prettierignore`:
  ```
  node_modules
  .next
  components/shadcn
  ```
- [ ] Update `/frontend/eslint.config.mjs` to ignore `components/shadcn`
- [ ] Create `/frontend/.oxlintignore`:
  ```
  node_modules
  .next
  components/shadcn
  ```
- [ ] Update `/frontend/package.json` scripts:
  ```json
  {
    "scripts": {
      "dev": "next dev",
      "build": "next build",
      "start": "next start",
      "lint": "oxlint . --fix && eslint . --fix",
      "format": "prettier --write .",
      "format:check": "prettier --check .",
      "tsc": "tsc --noEmit",
      "typegen": "next typegen"
    }
  }
  ```

```bash
git add .
git commit -m "Configure frontend linting and formatting"
```

---

## Step 23: Set Up Storybook

**Goal**: Initialize Storybook for component development.

- [ ] Initialize Storybook: `bunx storybook@latest init`
- [ ] Configure to work with Tailwind CSS
- [ ] Create `/frontend/stories/` directory
- [ ] Add example story for a shadcn Button component
- [ ] Update `/frontend/package.json` with storybook scripts if not added

```bash
git add .
git commit -m "Initialize Storybook"
```

---

## Step 24: Complete Docker Compose Setup

**Goal**: Finalize docker-compose.yml with all services.

- [ ] Update root `docker-compose.yml`:
  ```yaml
  version: '3.8'

  services:
    postgres:
      image: postgres:16
      environment:
        POSTGRES_USER: postgres
        POSTGRES_PASSWORD: postgres
        POSTGRES_DB: greenrats
      ports:
        - "5432:5432"
      volumes:
        - postgres_data:/var/lib/postgresql/data
      healthcheck:
        test: ["CMD-SHELL", "pg_isready -U postgres"]
        interval: 5s
        timeout: 5s
        retries: 5

    api:
      build:
        context: ./backend
        dockerfile: Dockerfile
        target: development
      ports:
        - "8000:8000"
      environment:
        - PORT=8000
        - DATABASE_URL=postgres://postgres:postgres@postgres:5432/greenrats?sslmode=disable
      depends_on:
        postgres:
          condition: service_healthy
      volumes:
        - ./backend:/app
        - /app/tmp

    ingest:
      build:
        context: ./backend
        dockerfile: Dockerfile
        target: ingest
      environment:
        - DATABASE_URL=postgres://postgres:postgres@postgres:5432/greenrats?sslmode=disable
      depends_on:
        postgres:
          condition: service_healthy
      volumes:
        - ./backend:/app

    frontend:
      build:
        context: ./frontend
        dockerfile: Dockerfile
        target: development
      ports:
        - "3000:3000"
      environment:
        - NEXT_PUBLIC_API_URL=http://localhost:8000
      depends_on:
        - api
      volumes:
        - ./frontend:/app
        - /app/node_modules
        - /app/.next

  volumes:
    postgres_data:
  ```
- [ ] Create `/backend/Dockerfile`:
  ```dockerfile
  FROM golang:1.23-alpine AS base
  WORKDIR /app
  RUN go install github.com/air-verse/air@latest
  COPY go.mod go.sum ./
  RUN go mod download

  FROM base AS development
  COPY . .
  CMD ["air", "-c", ".air.toml"]

  FROM base AS ingest
  COPY . .
  CMD ["air", "-c", ".air.ingest.toml"]

  FROM base AS build
  COPY . .
  RUN go build -o bin/api ./cmd/api
  RUN go build -o bin/ingest ./cmd/ingest

  FROM alpine:latest AS production
  WORKDIR /app
  COPY --from=build /app/bin/api .
  CMD ["./api"]
  ```
- [ ] Create `/frontend/Dockerfile`:
  ```dockerfile
  FROM oven/bun:latest AS base
  WORKDIR /app

  FROM base AS development
  COPY package.json bun.lockb ./
  RUN bun install
  COPY . .
  CMD ["bun", "dev"]

  FROM base AS build
  COPY package.json bun.lockb ./
  RUN bun install --frozen-lockfile
  COPY . .
  RUN bun run build

  FROM base AS production
  COPY --from=build /app/.next/standalone ./
  COPY --from=build /app/.next/static ./.next/static
  COPY --from=build /app/public ./public
  CMD ["bun", "run", "server.js"]
  ```

```bash
git add .
git commit -m "Complete Docker Compose setup with all services"
```

---

## Step 25: Create Basic Feature Scaffolding

**Goal**: Add minimal implementation for one feature to validate the architecture.

- [ ] Create `/backend/internal/features/tournaments/service.go`:
  - Create `TournamentService` struct
  - Add `List()` method that queries Ent
  - Add `GetByID()` method
- [ ] Create `/backend/internal/features/tournaments/handlers.go`:
  - Create `TournamentHandler` struct
  - Add `List` handler for `GET /api/v1/tournaments`
  - Add `Get` handler for `GET /api/v1/tournaments/:id`
- [ ] Create `/backend/internal/features/tournaments/types.go`:
  - Define response DTOs
- [ ] Register tournament routes in server
- [ ] Create `/frontend/features/tournaments/types.ts`:
  - Define Tournament type with Zod schema
- [ ] Create `/frontend/features/tournaments/queries.ts`:
  - Add `useTournaments` query hook
  - Add `useTournament` query hook
- [ ] Create basic `/frontend/app/(dashboard)/tournaments/page.tsx`:
  - Fetch and display tournaments list

```bash
git add .
git commit -m "Add tournaments feature scaffolding"
```

---

## Step 26: Update Root README

**Goal**: Finalize README with complete setup instructions.

- [ ] Update `/README.md` with:
  - Project description
  - Prerequisites (Go 1.23, Bun, Docker)
  - Quick start instructions
  - Environment setup guide
  - Development workflow
  - Available scripts/commands
  - Architecture overview (link to CLAUDE.md)

```bash
git add .
git commit -m "Update README with complete documentation"
```

---

## Step 27: Verify Full Stack

**Goal**: Ensure everything works together.

- [ ] Run `docker compose up --build`
- [ ] Verify PostgreSQL is healthy
- [ ] Verify API starts and `/health` returns OK
- [ ] Verify frontend starts and loads
- [ ] Verify Atlas migrations applied successfully
- [ ] Run backend tests: `go test ./...`
- [ ] Run frontend type check: `bun run tsc`
- [ ] Run linters on both

```bash
git add .
git commit -m "Verify full stack integration"
```

---

## Summary

After completing all steps, you should have:

- [x] Monorepo structure with backend and frontend
- [x] Go backend with Fiber, Ent, and Atlas migrations
- [x] Two executables: `api` and `ingest`
- [x] SSE support for live updates
- [x] WorkOS authentication on both ends
- [x] Next.js frontend with TanStack Query and shadcn/ui
- [x] Docker Compose for local development
- [x] Linting, formatting, and pre-commit hooks
- [x] One complete feature (tournaments) as a reference implementation

**Total commits**: 27
