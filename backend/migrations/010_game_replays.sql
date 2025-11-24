-- Migration 010: Game Replays and Anti-Cheat
-- Adds replay system and anti-cheat detection for all games

-- 1. Create game_replays table
CREATE TABLE IF NOT EXISTS game_replays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(255) NOT NULL UNIQUE,
    game_type VARCHAR(50) NOT NULL,
    players TEXT[] NOT NULL,
    moves JSONB NOT NULL DEFAULT '[]',
    states JSONB NOT NULL DEFAULT '[]',
    winner VARCHAR(255),
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    expires_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP + INTERVAL '30 days'),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_replays_session ON game_replays(session_id);
CREATE INDEX IF NOT EXISTS idx_replays_game_type ON game_replays(game_type);
CREATE INDEX IF NOT EXISTS idx_replays_winner ON game_replays(winner) WHERE winner IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_replays_expires ON game_replays(expires_at);

-- Create index for player lookups (GIN index for array contains)
CREATE INDEX IF NOT EXISTS idx_replays_players ON game_replays USING GIN(players);

-- 2. Create anticheat_alerts table
CREATE TABLE IF NOT EXISTS anticheat_alerts (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255),
    alert_type VARCHAR(50) NOT NULL CHECK (alert_type IN ('TIMING', 'WIN_RATE', 'INVALID_MOVE', 'STATISTICAL')),
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH')),
    details JSONB,
    reviewed BOOLEAN DEFAULT FALSE,
    reviewer_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alerts_user ON anticheat_alerts(user_id);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON anticheat_alerts(severity) WHERE severity = 'HIGH';
CREATE INDEX IF NOT EXISTS idx_alerts_unreviewed ON anticheat_alerts(reviewed) WHERE reviewed = FALSE;
CREATE INDEX IF NOT EXISTS idx_alerts_created ON anticheat_alerts(created_at DESC);

-- 3. Create useful views for monitoring

-- View: Recent suspicious activity
CREATE OR REPLACE VIEW recent_suspicious_activity AS
SELECT
    user_id,
    COUNT(*) as alert_count,
    COUNT(*) FILTER (WHERE severity = 'HIGH') as high_severity_count,
    COUNT(*) FILTER (WHERE severity = 'MEDIUM') as medium_severity_count,
    MAX(created_at) as last_alert,
    array_agg(DISTINCT alert_type) as alert_types
FROM anticheat_alerts
WHERE created_at > NOW() - INTERVAL '7 days'
GROUP BY user_id
HAVING COUNT(*) FILTER (WHERE severity = 'HIGH') > 0
ORDER BY high_severity_count DESC, alert_count DESC;

COMMENT ON VIEW recent_suspicious_activity IS 'Users with recent high-severity anti-cheat alerts';

-- View: Game replay statistics
CREATE OR REPLACE VIEW replay_statistics AS
SELECT
    game_type,
    COUNT(*) as total_games,
    COUNT(*) FILTER (WHERE completed_at IS NOT NULL) as completed_games,
    AVG(EXTRACT(EPOCH FROM (completed_at - started_at))) as avg_duration_seconds,
    COUNT(DISTINCT winner) as unique_winners
FROM game_replays
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY game_type;

COMMENT ON VIEW replay_statistics IS 'Game replay metrics over last 30 days';

-- 4. Create function to auto-delete expired replays
CREATE OR REPLACE FUNCTION delete_expired_replays()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM game_replays
    WHERE expires_at < NOW()
    RETURNING COUNT(*) INTO deleted_count;

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION delete_expired_replays IS 'Deletes replays older than 30 days. Run daily via cron.';

-- 5. Create function to check if user should be flagged
CREATE OR REPLACE FUNCTION should_flag_user(p_user_id VARCHAR(255))
RETURNS TABLE(should_flag BOOLEAN, reason TEXT, alert_count BIGINT) AS $$
BEGIN
    RETURN QUERY
    SELECT
        CASE
            WHEN COUNT(*) FILTER (WHERE severity = 'HIGH' AND created_at > NOW() - INTERVAL '7 days') >= 3
            THEN TRUE
            WHEN COUNT(*) FILTER (WHERE alert_type = 'WIN_RATE') >= 2
            THEN TRUE
            ELSE FALSE
        END as should_flag,
        CASE
            WHEN COUNT(*) FILTER (WHERE severity = 'HIGH' AND created_at > NOW() - INTERVAL '7 days') >= 3
            THEN '3+ high-severity alerts in 7 days'
            WHEN COUNT(*) FILTER (WHERE alert_type = 'WIN_RATE') >= 2
            THEN 'Multiple suspicious win rate alerts'
            ELSE 'No action needed'
        END as reason,
        COUNT(*) as alert_count
    FROM anticheat_alerts
    WHERE user_id = p_user_id
    AND created_at > NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION should_flag_user IS 'Determines if a user should be flagged for review based on anti-cheat alerts';

COMMENT ON TABLE game_replays IS 'Stores complete game replays for dispute resolution and anti-cheat verification';
COMMENT ON TABLE anticheat_alerts IS 'Logs all anti-cheat detection alerts for manual review';
