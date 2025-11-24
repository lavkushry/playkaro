# Architecture

This repo contains a Gin monolith that powers the end-to-end user journey plus a set of standalone services that mirror a production-grade layout (payment, match engine, casino engine, AI). Kong is provided as the edge proxy and Jaeger for traces.

## Runtime Modes
- **Monolith (default)** — `backend/main.go` serves REST + GraphQL + WebSocket on `:8080` against a single Postgres + Redis.
- **Microservices sandbox** — `backend/docker-compose.yml` runs payment-service, match-service, game-engine, ai-service, Postgres, Redis, Kafka, Kong (`:8000` proxy), and Jaeger (`:16686`). The frontend is currently wired to the monolith, not the gateway.

## Components
| Component | Port | Role | Dependencies |
| --- | --- | --- | --- |
| Gin API (`backend/`) | 8080 | Auth, wallet, betting, promotions, KYC, GraphQL, WebSocket | Postgres, Redis, (optional) wallet gRPC on 50051 |
| Payment Service | 8081 | Razorpay/Cashfree integration, payment ledger, webhooks | Postgres, Kafka (events), Kong |
| Match Service | 8082 | Sportsbook CRUD, odds simulation, odds WebSocket | Postgres, Redis, Kafka |
| Game Engine | 8083 | Casino/skill-game registry, sessions, WebSocket state | Redis |
| AI Service | 8084 | Recommendations and anti-cheat (FastAPI) | — |
| Kong Gateway | 8000/8001 | Routes traffic to services (`backend/kong.yml`) | Upstreams above |
| Infra | Postgres, Redis, Kafka, Jaeger | State, cache, events, tracing | — |

## Request Flows
- **Auth & Wallet**: REST under `/api/v1/auth` and `/api/v1/wallet`. JWT is HMAC via `JWT_SECRET`; wallets are created lazily. Transactions table records DEPOSIT/WITHDRAW/BET/WIN.
- **Sportsbook**: Matches live in Postgres. Odds updates are broadcast via Redis pub/sub to `ws://.../ws` using message type `odds_update`. Admin routes can create/update/settle matches; settling credits wallets for winning bets and writes transactions.
- **Casino / Seamless Wallet**: `GET /api/v1/casino/games` lists seeded games. Game launch returns a mock URL. Providers can call `/api/v1/game-wallet/{balance|debit|credit|rollback}` to run bets/wins/rollbacks and sync with `game_rounds`.
- **Payments**: `/api/v1/payment/deposit` writes `payment_transactions` and can short-circuit with `gateway=MOCK`. Webhooks (Razorpay-style) verify signatures when `RAZORPAY_WEBHOOK_SECRET` is set and credit the wallet.
- **Promotions & Referrals**: Bonuses table tracks amounts, wagering requirements, and expiry; referral codes pay out dual bonuses and update wallets.
- **WebSocket**: Gorilla-based hub broadcasts through Redis channel `broadcast_channel` for horizontal scale. Supported types: `odds_update`, `chat_message`.
- **GraphQL**: Schema at `backend/graph/schema.graphqls` exposes `me`, `balance`, `matches`, and basic auth/bet mutations.

## Frontend Surface
- React 19 + Vite + Tailwind. State via Zustand for auth, wallet, and bet slip; Apollo Client for GraphQL.
- Pages: Dashboard, Sportsbook, Admin, History, Payment, KYC, Casino, Promotions, Leaderboard, Analytics.
- API usage: REST for auth/wallet/bets/promotions, GraphQL for user/balance/matches, WebSocket for odds updates and chat.

## Data Model (monolith)
- **users, wallets, transactions** — core ledger with bonus balance column.
- **matches, bets** — sportsbook entities with potential winnings.
- **payment_transactions** — gateway/pay-out audit.
- **kyc_documents** — KYC status and reviewer.
- **games, game_sessions, game_rounds** — casino catalog and round ledger.
- **bonuses, referrals** — promotion tracking and referral bonuses.

## Observability & Ops
- Payment, match, and game-engine services emit OpenTelemetry traces (OTLP gRPC) to `otel-collector` alias (`jaeger` container in compose).
- Kong enables correlation IDs and Prometheus metrics via `backend/kong.yml`.
- Health endpoints: `/health` on every service; Jaeger UI at `http://localhost:16686`.
