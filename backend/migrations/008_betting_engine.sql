-- Migration 008: Betting Engine
-- Adds comprehensive betting system with optimistic locking and cash-out support

-- 1. Create bets table
CREATE TABLE IF NOT EXISTS bets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    match_id VARCHAR(255) NOT NULL,
    team VARCHAR(50) NOT NULL CHECK (team IN ('TEAM_A', 'TEAM_B', 'DRAW')),
    amount DECIMAL(10,2) NOT NULL CHECK (amount > 0),
    odds DECIMAL(5,2) NOT NULL CHECK (odds >= 1.01),
    potential_win DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'ACTIVE', 'SETTLED', 'REJECTED', 'CASHED_OUT')),
    result VARCHAR(20) CHECK (result IN ('WON', 'LOST', 'VOID')),
    cashed_out BOOLEAN DEFAULT FALSE,
    cash_out_at TIMESTAMP,
    cash_out_odds DECIMAL(5,2),
    cash_out_amount DECIMAL(10,2),
    settled_at TIMESTAMP,
    version INT DEFAULT 1 NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_bets_user ON bets(user_id);
CREATE INDEX IF NOT EXISTS idx_bets_match ON bets(match_id);
CREATE INDEX IF NOT EXISTS idx_bets_status ON bets(status);
CREATE INDEX IF NOT EXISTS idx_bets_created ON bets(created_at DESC);

-- Combined index for user's active bets
CREATE INDEX IF NOT EXISTS idx_bets_user_status ON bets(user_id, status) WHERE status IN ('PENDING', 'ACTIVE');

-- 3. Add version column to matches for optimistic locking
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name='matches' AND column_name='version') THEN
        ALTER TABLE matches ADD COLUMN version INT DEFAULT 1 NOT NULL;
    END IF;
END $$;

-- 4. Create useful views
CREATE OR REPLACE VIEW bet_statistics AS
SELECT
    user_id,
    COUNT(*) as total_bets,
    COUNT(*) FILTER (WHERE status = 'ACTIVE') as active_bets,
    COUNT(*) FILTER (WHERE result = 'WON') as won_bets,
    COUNT(*) FILTER (WHERE result = 'LOST') as lost_bets,
    SUM(amount) as total_wagered,
    SUM(CASE WHEN result = 'WON' THEN potential_win ELSE 0 END) as total_winnings,
    SUM(CASE WHEN result = 'WON' THEN potential_win ELSE 0 END) - SUM(amount) as net_profit
FROM bets
GROUP BY user_id;

COMMENT ON VIEW bet_statistics IS 'Aggregate betting statistics per user';

-- 5. Create function to auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to bets table
DROP TRIGGER IF EXISTS update_bets_updated_at ON bets;
CREATE TRIGGER update_bets_updated_at
    BEFORE UPDATE ON bets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Apply trigger to matches table
DROP TRIGGER IF EXISTS update_matches_updated_at ON matches;
CREATE TRIGGER update_matches_updated_at
    BEFORE UPDATE ON matches
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE bets IS 'Stores all user bets with support for live betting and cash-out';
