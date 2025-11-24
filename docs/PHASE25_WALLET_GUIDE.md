# Phase 25 — Enterprise Wallet Upgrade

This guide shows how to evolve the PlayKaro wallet into a production-grade, multi-balance ledger with bonus wagering, idempotency, KYC tiers, fraud checks, state machines, and reconciliation. It is written for the existing Gin/Postgres/Redis stack.

## SQL Migration
```sql
-- 01_split_wallets.sql
ALTER TABLE wallets
  DROP COLUMN IF EXISTS balance,
  DROP COLUMN IF EXISTS bonus,
  ADD COLUMN deposit_balance    DECIMAL(15,2) DEFAULT 0 NOT NULL,
  ADD COLUMN bonus_balance      DECIMAL(15,2) DEFAULT 0 NOT NULL,
  ADD COLUMN winnings_balance   DECIMAL(15,2) DEFAULT 0 NOT NULL,
  ADD COLUMN locked_balance     DECIMAL(15,2) DEFAULT 0 NOT NULL,
  ADD COLUMN currency           VARCHAR(3)    DEFAULT 'INR' NOT NULL,
  ADD COLUMN kyc_level          INT           DEFAULT 0 NOT NULL,
  ADD COLUMN daily_deposit_used DECIMAL(15,2) DEFAULT 0 NOT NULL,
  ADD COLUMN last_deposit_reset TIMESTAMP     DEFAULT NOW() NOT NULL,
  ADD COLUMN status             VARCHAR(20)   DEFAULT 'ACTIVE' NOT NULL;

-- Transactions now carry the balance bucket they affected
ALTER TABLE transactions
  ADD COLUMN IF NOT EXISTS bucket VARCHAR(20) DEFAULT 'deposit';
```

## Go Service (packages to add under `backend/internal/wallet`)
- Dependencies: `github.com/shopspring/decimal`, Redis client already present (`db.RDB`), logger of choice.
- Core types:
```go
type BalanceBucket string
const (
    BucketDeposit  BalanceBucket = "deposit"
    BucketBonus    BalanceBucket = "bonus"
    BucketWinnings BalanceBucket = "winnings"
    BucketLocked   BalanceBucket = "locked"
)

type Wallet struct { /* fields map to migration above */ }
```
- Service operations (all transactional with `SELECT ... FOR UPDATE`):
  - `GetWallet(ctx, userID uuid.UUID) (*Wallet, error)` — lazy create if missing.
  - `Credit(ctx, userID, amount decimal.Decimal, bucket BalanceBucket, meta Meta)` — writes `transactions` with `bucket`.
  - `LockForBet(ctx, userID, stake decimal.Decimal)` — apply deduction priority Bonus → Deposit → Winnings, move stake into `locked_balance`.
  - `SettleLocked(ctx, userID, stake, payout decimal.Decimal, win bool)` — release lock, credit winnings if win.
  - `ApplyWagering(ctx, userID, betAmount decimal.Decimal)` — iterate active bonuses and increment `wagered_amount`; unlock bonus when requirement met.
  - `ValidateDepositLimit(ctx, userID, amount)` — enforce KYC tier caps (see table below) and reset `daily_deposit_used` on day change.

### Idempotency (deposits/webhooks)
```go
// In PaymentService before wallet credit:
cacheKey := "idemp:" + req.TransactionID
if data, err := rdb.Get(ctx, cacheKey).Bytes(); err == nil { return decode(data) }
resp, err := process()
if err == nil {
    b, _ := json.Marshal(resp)
    rdb.Set(ctx, cacheKey, b, 24*time.Hour)
}
```

### KYC Tier Limits
```go
var KycLimits = map[int]struct{
    MaxDailyDeposit decimal.Decimal
    MaxWithdrawal   decimal.Decimal
    RequiredDocs    []string
}{
    0: {decimal.NewFromInt(10000),  decimal.Zero,                     nil},
    1: {decimal.NewFromInt(100000), decimal.NewFromInt(50000),        []string{"PAN"}},
    2: {decimal.NewFromInt(1000000),decimal.NewFromInt(500000),       []string{"PAN","AADHAAR","BANK"}},
}
```

### Fraud Checks (rules + unit tests)
Rules: (1) Velocity: >5 deposits/hour per user (`INCR key=fraud:velocity:<user>` with 1h TTL). (2) Unusual amount: >10× average deposit from DB. (3) IP churn: >3 IPs in 1h (`SET` with TTL). (4) New device + amount > 5000 (`SETNX device:<id>` with 30d TTL). Tests live in `backend/internal/fraud/fraud_test.go`, inject Redis test instance or use redismock, assert each rule triggers expected code.

### Transaction State Machine
`payment_transactions` keep `state`, `retries`, `last_error`. Processor transitions: `PENDING -> PROCESSING -> SETTLED/FAILED/REFUNDED`. Exponential backoff for retries. On SETTLED: call `wallet.Credit(..., BucketDeposit)`. On FAILED: do not credit; log and alert.

### Reconciliation
Cron every 5 minutes: fetch `payment_transactions` pending >10 minutes, query gateway, move to `SETTLED`/`FAILED`, and credit wallet accordingly. Log to tracing spans; emit metrics (`reconciliations_total`, `reconciliation_failures`).

## API Surface (monolith)
- `GET  /api/v1/wallet` → `{deposit_balance, bonus_balance, winnings_balance, locked_balance, currency, kyc_level, status}`
- `POST /api/v1/wallet/deposit` → `{amount, gateway, idempotency_key, device_id, ip}`; response `{transaction_id, state, new_balances}`
- `POST /api/v1/wallet/withdraw` → `{amount, destination}` enforcing KYC limit; response `{transaction_id, state}`
- `POST /api/v1/wallet/lock` (internal) → `{stake}` returns lock token for bet settlement.
- `POST /api/v1/wallet/settle` (internal) → `{lock_token, result: WIN/LOSE, payout}` updates locked/winnings.

## Redis Caching Strategy
- Cache available balance: `wallet:available:<user>` = float64 of `deposit + bonus + winnings - locked`; TTL 30s; refresh on mutations.
- Locks: `wallet:lock:<user>` via `SETNX` 5s around bet placement to prevent double-deduct.
- Idempotency: `idemp:<txn_id>` 24h.
- Fraud keys: `fraud:velocity:<user>`, `fraud:ip:<user>`, `fraud:device:<device>`.

## Payment Service Integration (existing `backend/services/payment-service`)
1. Accept `Idempotency-Key` header from gateway; use it as `idemp:<key>`.
2. Before charging, call wallet `ValidateDepositLimit`.
3. After gateway success, record `payment_transactions` state and invoke wallet `Credit(..., BucketDeposit)`; include `bucket` in `transactions`.
4. Webhook handler must be idempotent: short-circuit on cached response, verify signature, update state machine, credit wallet only once.
5. Emit OpenTelemetry spans around gateway calls and wallet writes.

## Unit Test Pointers
- Place tests in `backend/internal/{wallet,fraud}`; use `testing` + `github.com/stretchr/testify/require`.
- For fraud rules, mock Redis and repository averages; assert rule codes per scenario.
- For wallet deductions, simulate Bonus→Deposit→Winnings priority and locked balance behavior.
- For idempotency, call `ProcessDeposit` twice with same transaction ID; assert single DB write.
