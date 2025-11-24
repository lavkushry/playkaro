-- Migration 009: Live Odds Engine
-- Adds odds history tracking, market suspension, and monitoring views

-- 1. Add suspension fields to matches
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='matches' AND column_name='suspended') THEN
        ALTER TABLE matches ADD COLUMN suspended BOOLEAN DEFAULT FALSE;
        ALTER TABLE matches ADD COLUMN suspension_reason TEXT;
    END IF;
END $$;

-- 2. Create odds_history table for tracking all odds changes
CREATE TABLE IF NOT EXISTS odds_history (
    id SERIAL PRIMARY KEY,
    match_id VARCHAR(255) NOT NULL,
    odds_a DECIMAL(5,2) NOT NULL,
    odds_b DECIMAL(5,2) NOT NULL,
    odds_draw DECIMAL(5,2),
    total_bets INT DEFAULT 0,
    total_volume DECIMAL(12,2) DEFAULT 0,
    triggered_by VARCHAR(100), -- "AUTO", "ADMIN", "SIMULATOR", "Kelly Criterion adjustment"
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_odds_history_match ON odds_history(match_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_odds_history_time ON odds_history(created_at DESC);

-- 3. Create market_metrics view for real-time monitoring
CREATE OR REPLACE VIEW market_metrics AS
SELECT
    m.match_id,
    m.team_a,
    m.team_b,
    m.odds_a,
    m.odds_b,
    m.odds_draw,
    m.status,
    m.suspended,
    m.suspension_reason,
    COUNT(b.id) FILTER (WHERE b.status = 'ACTIVE') as active_bets,
    SUM(b.amount) FILTER (WHERE b.status = 'ACTIVE') as total_volume,
    SUM(CASE WHEN b.team = 'TEAM_A' AND b.status = 'ACTIVE' THEN b.potential_win ELSE 0 END) as liability_a,
    SUM(CASE WHEN b.team = 'TEAM_B' AND b.status = 'ACTIVE' THEN b.potential_win ELSE 0 END) as liability_b,
    SUM(CASE WHEN b.team = 'DRAW' AND b.status = 'ACTIVE' THEN b.potential_win ELSE 0 END) as liability_draw,
    SUM(CASE WHEN b.status = 'ACTIVE' THEN b.potential_win ELSE 0 END) as total_liability
FROM matches m
LEFT JOIN bets b ON m.match_id = b.match_id
WHERE m.status IN ('LIVE', 'UPCOMING')
GROUP BY m.match_id, m.team_a, m.team_b, m.odds_a, m.odds_b, m.odds_draw, m.status, m.suspended, m.suspension_reason;

COMMENT ON VIEW market_metrics IS 'Real-time market health metrics for risk management';

-- 4. Create odds_volatility view for monitoring rapid changes
CREATE OR REPLACE VIEW odds_volatility AS
SELECT
    oh1.match_id,
    oh1.odds_a as current_odds_a,
    oh1.odds_b as current_odds_b,
    oh2.odds_a as prev_odds_a,
    oh2.odds_b as prev_odds_b,
    ABS((oh1.odds_a - oh2.odds_a) / oh2.odds_a) as volatility_a,
    ABS((oh1.odds_b - oh2.odds_b) / oh2.odds_b) as volatility_b,
    oh1.created_at as latest_update,
    oh2.created_at as previous_update
FROM odds_history oh1
JOIN LATERAL (
    SELECT odds_a, odds_b, created_at
    FROM odds_history
    WHERE match_id = oh1.match_id AND id < oh1.id
    ORDER BY id DESC
    LIMIT 1
) oh2 ON true
WHERE oh1.created_at > NOW() - INTERVAL '10 minutes';

COMMENT ON VIEW odds_volatility IS 'Track rapid odds movements for suspension triggers';

-- 5. Create function to log odds changes automatically
CREATE OR REPLACE FUNCTION log_odds_change()
RETURNS TRIGGER AS $$
BEGIN
    -- Only log if odds actually changed
    IF (OLD.odds_a IS DISTINCT FROM NEW.odds_a) OR
       (OLD.odds_b IS DISTINCT FROM NEW.odds_b) OR
       (OLD.odds_draw IS DISTINCT FROM NEW.odds_draw) THEN

        INSERT INTO odds_history (match_id, odds_a, odds_b, odds_draw, triggered_by)
        VALUES (NEW.match_id, NEW.odds_a, NEW.odds_b, NEW.odds_draw, 'AUTO');
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to matches table
DROP TRIGGER IF EXISTS track_odds_changes ON matches;
CREATE TRIGGER track_odds_changes
    AFTER UPDATE ON matches
    FOR EACH ROW
    WHEN (OLD.odds_a IS DISTINCT FROM NEW.odds_a OR
          OLD.odds_b IS DISTINCT FROM NEW.odds_b OR
          OLD.odds_draw IS DISTINCT FROM NEW.odds_draw)
    EXECUTE FUNCTION log_odds_change();

-- 6. Add index for suspended markets
CREATE INDEX IF NOT EXISTS idx_matches_suspended ON matches(suspended) WHERE suspended = TRUE;

COMMENT ON TABLE odds_history IS 'Complete audit trail of all odds changes across all matches';
COMMENT ON COLUMN matches.suspended IS 'Market suspension flag - blocks new bets when TRUE';
COMMENT ON COLUMN matches.suspension_reason IS 'Human-readable reason for market suspension';
