# PlayKaro — Real Money Gaming Platform

PlayKaro is a Go + React real-money gaming platform with sportsbook, wallet, casino launcher, promotions, and admin tooling. The repo ships a Gin-based monolith plus optional microservices (payment, match, game engine, AI) and a Vite frontend.

## Repo Map
- `backend/` — Gin REST/GraphQL API with WebSocket hub, Postgres schema, Redis pub/sub, and gRPC wallet client stub.
- `backend/services/` — Experimental microservices: payment-service (`:8081`), match-service (`:8082`), game-engine (`:8083`), ai-service (`:8084`).
- `frontend/` — React 19 + Vite + Tailwind + Zustand + Apollo.
- `docker/` — Lightweight Postgres + Redis for local dev.
- `Memory/` — PRDs, UX specs, and project status notes.
- `docs/` — Architecture notes.

## Run Locally (monolith, recommended)
1. Start infrastructure (Postgres on `5433`, Redis on `6379`):
   ```bash
   cd docker
   docker-compose up -d
   ```
2. Configure the backend:
   ```bash
   cp backend/.env.example backend/.env
   # ensure DB_PORT matches the Postgres host port (5433 by default) and set JWT_SECRET
   ```
3. Run the API (Gin monolith):
   ```bash
   cd backend
   go run main.go
   ```
   - REST base: `http://localhost:8080/api/v1`
   - GraphQL: `POST http://localhost:8080/query` (Playground at `/playground`)
   - WebSocket: `ws://localhost:8080/ws` (`odds_update`, `chat_message`)
   - Database schema and seed data (matches, games) are created on first boot.
4. Run the frontend:
   ```bash
   cd frontend
   npm install
   npm run dev
   ```
   The UI assumes the API at `http://localhost:8080`; adjust `src/store/useAuthStore.js`, `useWalletStore.js`, `useBetStore.js`, and `src/apolloClient.js` if your backend URL differs.
5. Health check:
   ```bash
   curl http://localhost:8080/health
   ```

## Microservices Stack (optional sandbox)
- Spin up the service mesh:
  ```bash
  cd backend
  docker-compose up --build
  ```
- Gateway: `http://localhost:8000` (Kong). Upstreams: payment-service `:8081`, match-service `:8082` (`/ws/odds`), game-engine `:8083` (`/ws/sessions/:id`), ai-service `:8084`. Infra: Postgres, Redis, Kafka, Jaeger UI at `:16686`.
- The current frontend is wired to the monolith; use the gateway for manual service testing or integrations.

## Configuration
### Backend `.env`
| Variable | Purpose | Example |
| --- | --- | --- |
| `DB_HOST` | Postgres host | `localhost` |
| `DB_PORT` | Postgres port | `5433` (matches `docker/docker-compose.yml`) |
| `DB_USER` | Postgres user | `postgres` |
| `DB_PASSWORD` | Postgres password | `postgres` |
| `DB_NAME` | Database name | `playkaro` |
| `JWT_SECRET` | HMAC secret for JWT auth | **set a strong value** |
| `RAZORPAY_WEBHOOK_SECRET` | Optional webhook verification | *(blank in dev)* |

Notes: the gRPC wallet client expects a wallet service on `localhost:50051`; configure or stub accordingly. Redis defaults to `localhost:6379`.

### Microservices (via `backend/docker-compose.yml`)
- Payment service: `PAYMENT_DB_HOST`, `PAYMENT_DB_PASSWORD`, `RAZORPAY_KEY_ID`, `RAZORPAY_KEY_SECRET`, `PORT` (default `8081`).
- Match service: `MATCH_DB_HOST`, `MATCH_DB_PASSWORD`, `REDIS_URL`, `PORT` (default `8082`).
- Game engine: `PORT` (default `8083`); mock auth via `X-User-ID`.
- AI service: Python FastAPI on `8084`.

## API Surface (monolith)
- Auth: `POST /api/v1/auth/register`, `POST /api/v1/auth/login`.
- Wallet: `GET /wallet/`, `POST /wallet/deposit`, `POST /wallet/withdraw`.
- Betting: `GET /matches`, `POST /bet/`.
- Admin: `POST /admin/matches`, `PUT /admin/matches/:id/odds`, `POST /admin/matches/:id/settle`, `POST /admin/kyc/approve`.
- Payments: `POST /payment/deposit`, `POST /payment/withdraw`, `POST /payment/webhook/razorpay`.
- KYC: `POST /kyc/upload`, `GET /kyc/status`.
- Casino/seamless wallet: `GET /casino/games`, `GET /casino/launch`, `POST /game-wallet/balance|debit|credit|rollback`.
- Promotions: `GET /promotions/bonuses`, `POST /promotions/claim`, `POST /promotions/referral/generate`, `POST /promotions/referral/apply`, `GET /promotions/leaderboard`.
- History: `GET /transactions`, `GET /bets`.
- WebSocket: `ws://localhost:8080/ws` (odds + chat).
- GraphQL: `/query` with schema in `backend/graph/schema.graphqls`.

## Seed Data
- A demo cricket match is inserted if none exist; casino games are seeded with thumbnails and RTP values.
- Wallets are lazily created on first balance lookup or deposit.

## Testing
- Manual checks: register, login, deposit, place a bet, view history, settle a match as admin, verify wallet credit.
- A simple end-to-end harness lives in `backend/tests/e2e` (expects running services; adjust auth headers as needed).

## More Docs
- Architecture notes: `docs/ARCHITECTURE.md`
- PRDs and UX specs: `Memory/`
