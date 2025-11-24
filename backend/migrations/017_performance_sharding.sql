-- Migration 017: Performance Optimization (Sharding)
-- Implements native partitioning for high-volume tables

-- Note: In a real production migration, we would migrate data from old tables to new partitioned tables.
-- Here we define the schema for new partitioned tables.

-- 1. Partitioned Ledger (Monthly)
CREATE TABLE IF NOT EXISTS ledger_partitioned (
    id UUID,
    transaction_id VARCHAR(255),
    user_id VARCHAR(255),
    type VARCHAR(50),
    amount DECIMAL(12,2),
    balance_type VARCHAR(50),
    reference_id VARCHAR(255),
    reference_type VARCHAR(50),
    balance_after DECIMAL(12,2),
    state VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMP NOT NULL
) PARTITION BY RANGE (created_at);

-- Create partitions for next 12 months
CREATE TABLE ledger_y2025m11 PARTITION OF ledger_partitioned
    FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
CREATE TABLE ledger_y2025m12 PARTITION OF ledger_partitioned
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');
CREATE TABLE ledger_y2026m01 PARTITION OF ledger_partitioned
    FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

CREATE INDEX IF NOT EXISTS idx_ledger_part_user ON ledger_partitioned(user_id);
CREATE INDEX IF NOT EXISTS idx_ledger_part_created ON ledger_partitioned(created_at);

-- 2. Partitioned Bets (Monthly)
CREATE TABLE IF NOT EXISTS bets_partitioned (
    id UUID,
    user_id VARCHAR(255),
    match_id VARCHAR(255),
    amount DECIMAL(12,2),
    odds DECIMAL(10,2),
    potential_payout DECIMAL(12,2),
    status VARCHAR(50),
    created_at TIMESTAMP NOT NULL
) PARTITION BY RANGE (created_at);

CREATE TABLE bets_y2025m11 PARTITION OF bets_partitioned
    FOR VALUES FROM ('2025-11-01') TO ('2025-12-01');
CREATE TABLE bets_y2025m12 PARTITION OF bets_partitioned
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');

CREATE INDEX IF NOT EXISTS idx_bets_part_user ON bets_partitioned(user_id);
CREATE INDEX IF NOT EXISTS idx_bets_part_match ON bets_partitioned(match_id);

-- 3. Partitioned Analytics Events (Daily)
CREATE TABLE IF NOT EXISTS analytics_events_partitioned (
    id UUID,
    user_id VARCHAR(255),
    event_type VARCHAR(50),
    event_data JSONB,
    created_at TIMESTAMP NOT NULL
) PARTITION BY RANGE (created_at);

-- Create partitions for next few days (in prod, use pg_partman)
CREATE TABLE analytics_y2025m11d24 PARTITION OF analytics_events_partitioned
    FOR VALUES FROM ('2025-11-24') TO ('2025-11-25');
CREATE TABLE analytics_y2025m11d25 PARTITION OF analytics_events_partitioned
    FOR VALUES FROM ('2025-11-25') TO ('2025-11-26');

CREATE INDEX IF NOT EXISTS idx_analytics_part_time ON analytics_events_partitioned(created_at);

COMMENT ON TABLE ledger_partitioned IS 'Monthly partitioned ledger for high-volume transactions';
