# GreenRats

A golf pick'em website where users pick one golfer per PGA Tour/major tournament throughout the season.

## Overview

GreenRats is a monorepo containing:
- **Backend** (`/backend`): Go services with Fiber (HTTP) and Ent (ORM)
- **Frontend** (`/frontend`): Next.js 16 App Router with React 19

## Quick Start

```bash
# Start the full stack with Docker
docker compose up --build

# View the application
# Frontend: http://localhost:3000
# Backend:  http://localhost:8000
# Health:   http://localhost:8000/health
```

## Local Development

### Prerequisites

- Go 1.25+
- Bun 1.x
- Docker & Docker Compose
- PostgreSQL 16 (via Docker or local)

### Backend

```bash
cd backend

# Install dependencies
go mod download

# Start with hot reload
air                           # API server
air -c .air.ingest.toml       # Ingest service

# Run tests
go test ./...

# Lint
golangci-lint run
```

### Frontend

```bash
cd frontend

# Install dependencies
bun install

# Start dev server
bun dev

# Build
bun run build

# Lint & format
bun run lint
bun run format

# Type check
bun run tsc

# Storybook
bun run storybook
```

### Database

```bash
# Start PostgreSQL
docker compose up postgres -d

# Generate Ent code (after schema changes)
cd backend && go generate ./ent

# Create migration
atlas migrate diff <name> --env local

# Apply migrations
atlas migrate apply --env local
```

## Core Mechanics

- Users pick one golfer per tournament
- Once a golfer is picked, they cannot be reused for the rest of the season
- Leaderboards track cumulative earnings across all tournaments
- Users compete within leagues

## Architecture

### Backend Services

- **API** (`cmd/api`): HTTP server handling user requests, authentication, SSE
- **Ingest** (`cmd/ingest`): Background service fetching external golf data

### Tech Stack

| Layer | Technology |
|-------|------------|
| HTTP | Fiber v2 |
| ORM | Ent |
| Migrations | Atlas |
| Config | Viper |
| Auth | WorkOS JWT |
| Real-time | SSE |

### Frontend Stack

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

## Environment Variables

### Backend (`backend/.env`)

```env
PORT=8000
ENV=development
DATABASE_URL=postgres://greenrats:greenrats@localhost:5432/greenrats?sslmode=disable
WORKOS_API_KEY=
WORKOS_CLIENT_ID=
SCRATCHGOLF_API_KEY=
BALLDONTLIE_API_KEY=
```

### Frontend (`frontend/.env.local`)

```env
WORKOS_CLIENT_ID=
WORKOS_API_KEY=
WORKOS_REDIRECT_URI=http://localhost:3000/callback
WORKOS_COOKIE_PASSWORD=  # 32+ characters
NEXT_PUBLIC_API_URL=http://localhost:8000
```

## Deployment (Railway)

GreenRats is deployed on [Railway](https://railway.app) with two environments:

### Environments

| Environment | Purpose | Email Sending |
|-------------|---------|---------------|
| Development | Staging/testing | Disabled |
| Production | Live application | Enabled |

### Services

Each environment runs three services:

1. **API** - Fiber HTTP server (`backend/cmd/api`)
2. **Ingest** - Background data sync service (`backend/cmd/ingest`)
3. **Frontend** - Next.js application

Plus a **PostgreSQL** database addon.

### Railway Config Paths

Configure each Railway service to use the matching config file:

**API**
```
backend/railway/api.json
```

**Ingest**
```
backend/railway/ingest.json
```

**Frontend**
```
frontend/railway/web.json
```

### Deployment Workflow

1. **Push to main** triggers automatic deployment to Development
2. **Create release** or manual promote deploys to Production

### Environment Variables

Configure these in Railway for each environment:

**Backend (API & Ingest)**
```
ENV=production|development
DATABASE_URL=<railway-provided>
WORKOS_API_KEY=
WORKOS_CLIENT_ID=
SCRATCH_GOLF_API_KEY=
BALL_DONT_LIE_API_KEY=
SENTRY_DSN=
RESEND_API_KEY=
FROM_EMAIL=noreply@greenrats.com
SEND_EMAILS=true|false
```

**Frontend**
```
WORKOS_CLIENT_ID=
WORKOS_API_KEY=
WORKOS_REDIRECT_URI=https://greenrats.com/api/auth/callback
WORKOS_COOKIE_PASSWORD=
NEXT_PUBLIC_API_URL=https://api.greenrats.com
NEXT_PUBLIC_SENTRY_DSN=
```

### Manual Deployment

```bash
# Install Railway CLI
npm install -g @railway/cli

# Login
railway login

# Link to project
railway link

# Deploy specific service
railway up --service api
railway up --service frontend
```

## Development Guidelines

See [CLAUDE.md](./CLAUDE.md) for detailed development instructions, coding conventions, and workflow guidance.

## License

Private - All rights reserved.
