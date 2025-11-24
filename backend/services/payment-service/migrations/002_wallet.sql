-- Wallet and Ledger Schema
-- Database: payments_db

-- Wallets table (Current State)
CREATE TABLE IF NOT EXISTS wallets (
    user_id UUID PRIMARY KEY,
    balance DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'INR',
    updated_at TIMESTAMP DEFAULT NOW(),
    version INT DEFAULT 1 -- Optimistic Locking
);

-- Ledger table (History/Audit Trail)
CREATE TABLE IF NOT EXISTS ledger (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id VARCHAR(100) UNIQUE NOT NULL, -- Idempotency Key
    user_id UUID NOT NULL REFERENCES wallets(user_id),
    type VARCHAR(50) NOT NULL,
    amount DECIMAL(15, 2) NOT NULL,
    reference_id VARCHAR(100),
    reference_type VARCHAR(50),
    balance_after DECIMAL(15, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ledger_user_id ON ledger(user_id);
CREATE INDEX IF NOT EXISTS idx_ledger_ref ON ledger(reference_id, reference_type);
