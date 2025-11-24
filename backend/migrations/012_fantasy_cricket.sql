-- Migration 012: Fantasy Cricket
-- Stores fantasy teams, contests, and scoring data

-- 1. Create fantasy_contests table
CREATE TABLE IF NOT EXISTS fantasy_contests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    entry_fee DECIMAL(10,2) NOT NULL,
    prize_pool DECIMAL(12,2) NOT NULL,
    max_teams INT NOT NULL,
    current_teams INT DEFAULT 0,
    status VARCHAR(50) NOT NULL, -- OPEN, LIVE, COMPLETED, CANCELLED
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fantasy_contests_match ON fantasy_contests(match_id);
CREATE INDEX IF NOT EXISTS idx_fantasy_contests_status ON fantasy_contests(status);

-- 2. Create fantasy_teams table
CREATE TABLE IF NOT EXISTS fantasy_teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    contest_id UUID REFERENCES fantasy_contests(id),
    match_id VARCHAR(255) NOT NULL,
    players JSONB NOT NULL, -- Array of FantasyPlayer
    captain_id VARCHAR(255) NOT NULL,
    vice_captain_id VARCHAR(255) NOT NULL,
    total_cost DECIMAL(5,2) NOT NULL,
    total_points DECIMAL(8,2) DEFAULT 0,
    rank INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fantasy_teams_user ON fantasy_teams(user_id);
CREATE INDEX IF NOT EXISTS idx_fantasy_teams_contest ON fantasy_teams(contest_id);
CREATE INDEX IF NOT EXISTS idx_fantasy_teams_match ON fantasy_teams(match_id);

-- 3. Create fantasy_points_history table (optional, for detailed breakdown)
CREATE TABLE IF NOT EXISTS fantasy_points_history (
    id SERIAL PRIMARY KEY,
    team_id UUID REFERENCES fantasy_teams(id),
    player_id VARCHAR(255) NOT NULL,
    points DECIMAL(8,2) NOT NULL,
    breakdown JSONB, -- {runs: 10, wickets: 25, ...}
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fantasy_points_team ON fantasy_points_history(team_id);

-- 4. Create view for contest leaderboards
CREATE OR REPLACE VIEW contest_leaderboard AS
SELECT
    t.contest_id,
    t.id as team_id,
    t.user_id,
    t.total_points,
    RANK() OVER (PARTITION BY t.contest_id ORDER BY t.total_points DESC) as rank
FROM fantasy_teams t;

COMMENT ON TABLE fantasy_contests IS 'Fantasy cricket contests for specific matches';
COMMENT ON TABLE fantasy_teams IS 'User created fantasy teams for contests';
