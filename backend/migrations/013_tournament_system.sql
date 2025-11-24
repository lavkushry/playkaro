-- Migration 013: Tournament System
-- Stores tournament configuration, participants, and matches

-- 1. Create tournaments table
CREATE TABLE IF NOT EXISTS tournaments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    game_type VARCHAR(50) NOT NULL,
    entry_fee DECIMAL(10,2) NOT NULL,
    prize_pool DECIMAL(12,2) NOT NULL,
    max_players INT NOT NULL,
    current_players INT DEFAULT 0,
    status VARCHAR(50) NOT NULL, -- REGISTRATION, ACTIVE, COMPLETED, CANCELLED
    config JSONB NOT NULL, -- Bracket type, prize rules
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tournaments_status ON tournaments(status);
CREATE INDEX IF NOT EXISTS idx_tournaments_game_type ON tournaments(game_type);

-- 2. Create tournament_participants table
CREATE TABLE IF NOT EXISTS tournament_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID REFERENCES tournaments(id),
    user_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL, -- REGISTERED, ELIMINATED, WINNER
    rank INT,
    prize_amount DECIMAL(12,2) DEFAULT 0,
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tournament_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_participants_tournament ON tournament_participants(tournament_id);
CREATE INDEX IF NOT EXISTS idx_participants_user ON tournament_participants(user_id);

-- 3. Create tournament_matches table
CREATE TABLE IF NOT EXISTS tournament_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID REFERENCES tournaments(id),
    round INT NOT NULL,
    match_index INT NOT NULL,
    player_1_id VARCHAR(255),
    player_2_id VARCHAR(255),
    winner_id VARCHAR(255),
    status VARCHAR(50) NOT NULL, -- SCHEDULED, IN_PROGRESS, COMPLETED
    next_match_id UUID REFERENCES tournament_matches(id),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_matches_tournament ON tournament_matches(tournament_id);
CREATE INDEX IF NOT EXISTS idx_matches_round ON tournament_matches(tournament_id, round);

COMMENT ON TABLE tournaments IS 'Competitive tournaments with bracket system';
COMMENT ON TABLE tournament_participants IS 'Users registered for tournaments';
COMMENT ON TABLE tournament_matches IS 'Individual matches within a tournament bracket';
