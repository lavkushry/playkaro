-- Migration 007: Multi-Currency Wallet System
-- This migration adds support for split balances (Deposit/Bonus/Winnings)
-- and enhances the transaction pipeline with states

-- 1. Add new balance columns to wallets table
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='wallets' AND column_name='deposit_balance') THEN
        ALTER TABLE wallets ADD COLUMN deposit_balance DECIMAL(10,2) DEFAULT 0;
        ALTER TABLE wallets ADD COLUMN bonus_balance DECIMAL(10,2) DEFAULT 0;
        ALTER TABLE wallets ADD COLUMN winnings_balance DECIMAL(10,2) DEFAULT 0;

        -- Migrate existing balance to deposit_balance
        UPDATE wallets SET deposit_balance = balance WHERE deposit_balance = 0;
    END IF;
END $$;

-- 2. Create bonuses table
CREATE TABLE IF NOT EXISTS bonuses (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'ACTIVE',
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bonuses_user_id ON bonuses(user_id);
CREATE INDEX IF NOT EXISTS idx_bonuses_expires_at ON bonuses(expires_at) WHERE status = 'ACTIVE';

-- 3. Add transaction state and metadata to ledger
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='ledger' AND column_name='state') THEN
        ALTER TABLE ledger ADD COLUMN state VARCHAR(50) DEFAULT 'SETTLED';
        ALTER TABLE ledger ADD COLUMN balance_type VARCHAR(50) DEFAULT 'DEPOSIT';
        ALTER TABLE ledger ADD COLUMN metadata JSONB;
    END IF;
END $$;

-- 4. Create device_fingerprints table for fraud detection
CREATE TABLE IF NOT EXISTS device_fingerprints (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    device_hash VARCHAR(255) NOT NULL,
    ip_address VARCHAR(50),
    user_agent TEXT,
    first_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, device_hash)
);

CREATE INDEX IF NOT EXISTS idx_device_fingerprints_hash ON device_fingerprints(device_hash);
CREATE INDEX IF NOT EXISTS idx_device_fingerprints_user ON device_fingerprints(user_id);

-- 5. Add some helpful views
CREATE OR REPLACE VIEW wallet_health AS
SELECT
    user_id,
    balance as total_balance,
    deposit_balance,
    bonus_balance,
    winnings_balance,
    (deposit_balance + bonus_balance + winnings_balance) as calculated_balance,
    ABS(balance - (deposit_balance + bonus_balance + winnings_balance)) as balance_mismatch,
    CASE
        WHEN ABS(balance - (deposit_balance + bonus_balance + winnings_balance)) > 0.01
        THEN 'MISMATCH'
        ELSE 'HEALTHY'
    END as health_status
FROM wallets;

COMMENT ON VIEW wallet_health IS 'Monitor wallet balance integrity';
