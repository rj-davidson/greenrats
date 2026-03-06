# GreenRats

A golf pick'em website where users pick one golfer per PGA Tour/major tournament throughout the season.

## Core Mechanics

- Users pick one golfer per tournament
- Once a golfer is picked, they cannot be reused for the rest of the season
- Leaderboards track cumulative earnings across all tournaments
- Users compete within leagues

## Quick Start

```bash
# Start the full stack with Docker
docker compose up --build

# Frontend: http://localhost:3000
# Backend:  http://localhost:8000
```

## Prerequisites

- Go 1.25+
- Bun 1.x
- Docker & Docker Compose
- PostgreSQL 16 (via Docker or local)

## Local Development

### Backend

```bash
cd backend
go mod download
air                           # API server with hot reload
air -c .air.ingest.toml       # Ingest service with hot reload
go test ./...                 # Run tests
golangci-lint run             # Lint
```

### Frontend

```bash
cd frontend
bun install
bun dev                       # Dev server
bun run build                 # Production build
bun run lint                  # Lint
bun run format                # Format
bun run tsc                   # Type check
bun run storybook             # Storybook
```

### Database

```bash
cd backend
go generate ./ent                       # Regenerate Ent code after schema changes
atlas migrate diff <name> --env local   # Create migration
atlas migrate apply --env local         # Apply migrations
```

### Environment Variables

Copy the example files and fill in your values:

- Backend: `backend/.env` (see `backend/.env.example`)
- Frontend: `frontend/.env.local` (see `frontend/.env.local.example`)

## Tech Stack

### Backend

| Layer | Technology |
|-------|------------|
| HTTP | Fiber v2 |
| ORM | Ent |
| Migrations | Atlas |
| Auth | WorkOS JWT |
| Real-time | SSE |

### Frontend

| Layer | Technology |
|-------|------------|
| Framework | Next.js 16 (App Router) |
| UI | React 19, Tailwind CSS v4, shadcn/ui |
| State | TanStack Query v5 |
| Forms | React Hook Form, Zod |
| Auth | WorkOS AuthKit |

## Project Structure

```
greenrats/
├── backend/
│   ├── cmd/
│   │   ├── api/          # API server entrypoint
│   │   └── ingest/       # Ingest service entrypoint
│   ├── internal/
│   │   ├── config/       # Configuration
│   │   ├── server/       # Fiber setup, routes, middleware
│   │   ├── auth/         # WorkOS JWT middleware
│   │   ├── sse/          # Server-Sent Events
│   │   ├── features/     # Domain features (tournaments, etc.)
│   │   └── external/     # External API clients
│   ├── ent/              # Ent schemas and generated code
│   └── migrations/       # Atlas migrations
│
├── frontend/
│   ├── app/              # Next.js App Router pages
│   ├── components/       # UI components
│   │   ├── shadcn/       # Generated shadcn components
│   │   └── core/         # App-wide shared components
│   ├── features/         # Feature modules
│   ├── lib/              # Utilities and providers
│   └── stories/          # Storybook stories
│
└── docker-compose.yml    # Full stack orchestration
```

## License

This project is licensed under the [PolyForm Noncommercial License 1.0.0](./LICENSE) with additional restrictions prohibiting gambling and resale use. You are free to use, modify, and share this software for non-commercial purposes only.
