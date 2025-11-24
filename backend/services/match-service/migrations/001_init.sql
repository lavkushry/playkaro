-- Match Service Database Schema
-- Database: matches_db

-- Matches table
CREATE TABLE IF NOT EXISTS matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id VARCHAR(100) UNIQUE NOT NULL,
    sport VARCHAR(50) NOT NULL,
    team_a VARCHAR(255) NOT NULL,
    team_b VARCHAR(255) NOT NULL,
    odds_a DECIMAL(5, 2) NOT NULL,
    odds_b DECIMAL(5, 2) NOT NULL,
    odds_draw DECIMAL(5, 2) DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'UPCOMING',
    start_time TIMESTAMP NOT NULL,
    league VARCHAR(255),
    venue VARCHAR(255),
    result VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    settled_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_matches_status ON matches(status);
CREATE INDEX IF NOT EXISTS idx_matches_start_time ON matches(start_time);
CREATE INDEX IF NOT EXISTS idx_matches_sport ON matches(sport);
CREATE INDEX IF NOT EXISTS idx_matches_match_id ON matches(match_id);

-- Odds history table
CREATE TABLE IF NOT EXISTS odds_history (
    id SERIAL PRIMARY KEY,
    match_id UUID REFERENCES matches(id),
    odds_a DECIMAL(5, 2),
    odds_b DECIMAL(5, 2),
    odds_draw DECIMAL(5, 2),
    timestamp TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_odds_history_match_id ON odds_history(match_id);
CREATE INDEX IF NOT EXISTS idx_odds_history_timestamp ON odds_history(timestamp);

-- Markets table (for advanced betting)
CREATE TABLE IF NOT EXISTS markets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID REFERENCES matches(id),
    market_type VARCHAR(100) NOT NULL,
    market_name VARCHAR(255),
    status VARCHAR(50) DEFAULT 'ACTIVE',
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_markets_match_id ON markets(match_id);
CREATE INDEX IF NOT EXISTS idx_markets_status ON markets(status);
