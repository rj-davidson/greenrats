# Repository Guidelines

## Project Structure & Module Organization
- `backend/` contains Go services (Fiber + Ent). Entrypoints live in `backend/cmd/api` and `backend/cmd/ingest`. Feature modules are under `backend/internal/features`, and Ent schemas/generated code live in `backend/ent` (do not edit generated files).
- `frontend/` is a Next.js 16 App Router app. Pages are in `frontend/app`, shared UI is in `frontend/components`, and feature modules are in `frontend/features`.
- Database migrations are in `backend/migrations`, with Atlas configuration in `backend/atlas.hcl`.

## Build, Test, and Development Commands
- `docker compose up --build` runs the full stack (frontend on `:3000`, backend on `:8000`).
- Backend (from `backend/`):
  - `go mod download` installs dependencies.
  - `air` runs the API with hot reload; `air -c .air.ingest.toml` runs the ingest service.
  - `go test ./...` runs all backend tests.
  - `golangci-lint run` runs linting.
- Frontend (from `frontend/`):
  - `bun install` installs dependencies.
  - `bun dev` starts the dev server.
  - `bun run build` builds production assets.
  - `bun run lint` and `bun run format` run linting and formatting.
  - `bun run storybook` starts Storybook.

## Coding Style & Naming Conventions
- Go: use `gofmt`/`goimports`, idiomatic Go, and wrap errors with context (`fmt.Errorf("...: %w", err)`).
- TypeScript/React: follow ESLint + oxlint rules; format with `oxfmt`.
- Keep feature modules cohesive: handlers delegate to services; request/response types stay in `types` files.
- Avoid code comments unless the code is explicitly non-idiomatic.

## Testing Guidelines
- Backend: `go test` with `testify` for assertions; name files `*_test.go`.
- Frontend: Storybook is the primary UI validation tool; Vitest is available but no default test script is defined yet.

## Commit & Pull Request Guidelines
- Commits use short, imperative, sentence-case summaries (e.g., `Fix AuthKit redirect URL`).
- PRs should include a clear description, linked issues, and screenshots for UI changes. Include migration notes when touching `backend/migrations`.

## Additional Notes
- See `CLAUDE.md` for deeper architecture and workflow details.
