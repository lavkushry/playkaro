# Repository Guidelines

## Project Structure
- `backend/`: Gin monolith (REST, GraphQL, WebSocket), Postgres/Redis setup, gRPC wallet stub, plus microservices under `backend/services/`.
- `frontend/`: React 19 + Vite + Tailwind UI; Zustand stores and Apollo client live in `src/`.
- `docker/`: Local Postgres (`5433`) and Redis compose file.
- `docs/`: Architecture notes. `Memory/`: PRDs and UX specs.
- Tests: `backend/tests/e2e` contains a simple end-to-end harness; service-specific tests can be added alongside packages.

## Build, Run, and Test
- Monolith API: `cd backend && go run main.go` (requires Postgres/Redis; copy `.env.example`).
- Frontend: `cd frontend && npm install && npm run dev` (default `:5173`), `npm run build` for production bundle, `npm run lint` for ESLint.
- Local infra: `cd docker && docker-compose up -d` (Postgres on `5433`, Redis on `6379`).
- Microservices sandbox: `cd backend && docker-compose up --build` (Kong gateway `:8000`, Jaeger `:16686`).
- Go tests (add as needed): `cd backend && go test ./...`; E2E harness: `cd backend/tests/e2e && go run main.go` (expects running services).

## Coding Style & Naming
- Go: follow `gofmt` (tabs for indent, camelCase identifiers). Keep handlers/middleware small and composable.
- JS/React: 2-space indent, JSX components in `PascalCase`, hooks/stores in `camelCase`. Run `npm run lint`.
- Filenames: Go packages lowercase with underscores avoided; React components `ComponentName.jsx`.
- Config: `.env` in `backend/` for API; frontend endpoints are currently hard-coded (update `src/store/*` and `src/apolloClient.js` if hosts change).

## Testing Guidelines
- Prefer `go test` per package; place fixtures near tests. Name files `*_test.go`.
- For UI changes, add lightweight checks or stories where feasible; ensure lint passes.
- E2E script is illustrative; adapt headers/tokens before relying on it for CI.

## Commit & PR Process
- Commits: use clear, imperative subjects (e.g., `Add wallet bonus calculation`, `Fix match odds broadcast`). Group related changes; avoid noisy WIP messages.
- Pull Requests: include a brief summary, testing done (`go test`, `npm run lint`, manual steps), affected routes/areas, and screenshots for UI changes. Link issues/PRDs when relevant.

## Security & Configuration
- Set strong `JWT_SECRET`; never commit real Razorpay keys. Use `gateway=MOCK` for payments in dev.
- Database ports differ between `docker/` (5433) and `backend/docker-compose.yml` (internal 5432); align envs accordingly.
