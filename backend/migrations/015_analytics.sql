-- Migration 015: Analytics Service
-- Stores raw events and aggregated metrics

-- 1. Create analytics_events table (Partitioned by month ideally, but simple here)
CREATE TABLE IF NOT EXISTS analytics_events (
    id UUID PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_events_user ON analytics_events(user_id);
CREATE INDEX IF NOT EXISTS idx_events_type ON analytics_events(event_type);
CREATE INDEX IF NOT EXISTS idx_events_time ON analytics_events(created_at);

-- 2. Create user_metrics table (Daily Aggregates)
CREATE TABLE IF NOT EXISTS user_metrics (
    user_id VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    total_deposits DECIMAL(12,2) DEFAULT 0,
    total_withdrawals DECIMAL(12,2) DEFAULT 0,
    total_wagers DECIMAL(12,2) DEFAULT 0,
    total_payouts DECIMAL(12,2) DEFAULT 0,
    session_count INT DEFAULT 0,
    last_active TIMESTAMP,
    PRIMARY KEY (user_id, date)
);

-- 3. Create game_metrics table (Daily Aggregates)
CREATE TABLE IF NOT EXISTS game_metrics (
    game_type VARCHAR(50) NOT NULL,
    date DATE NOT NULL,
    total_rounds INT DEFAULT 0,
    total_wagers DECIMAL(12,2) DEFAULT 0,
    total_payouts DECIMAL(12,2) DEFAULT 0,
    unique_players INT DEFAULT 0,
    PRIMARY KEY (game_type, date)
);

COMMENT ON TABLE analytics_events IS 'Raw event log for all system activities';
COMMENT ON TABLE user_metrics IS 'Daily aggregated stats per user';
COMMENT ON TABLE game_metrics IS 'Daily aggregated stats per game type';
