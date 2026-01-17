# GreenRats

A golf pick'em website where users pick one golfer per PGA Tour/major tournament throughout the season.

## Overview

GreenRats is a monorepo containing:
- **Backend** (`/backend`): Go services with Fiber (HTTP) and Ent (ORM)
- **Frontend** (`/frontend`): Next.js App Router with React

## Quick Start

```bash
# Start the full stack
docker compose up --build

# View the application
# Frontend: http://localhost:3000
# Backend:  http://localhost:8000
```

## Core Mechanics

- Users pick one golfer per tournament
- Once a golfer is picked, they cannot be reused for the rest of the season
- Leaderboards track cumulative earnings across all tournaments
- Users compete within leagues

## Development

See [CLAUDE.md](./CLAUDE.md) for detailed development instructions.

## License

Private - All rights reserved.
