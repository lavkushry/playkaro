-- Migration 011: Teen Patti Game
-- Stores game state and player history for Teen Patti

-- 1. Create teen_patti_games table
CREATE TABLE IF NOT EXISTS teen_patti_games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_id VARCHAR(255) NOT NULL,
    state VARCHAR(50) NOT NULL, -- WAITING, DEALING, BETTING, SHOWDOWN, FINISHED
    pot_amount DECIMAL(12,2) DEFAULT 0,
    boot_amount DECIMAL(10,2) NOT NULL,
    current_stake DECIMAL(10,2) NOT NULL,
    current_turn VARCHAR(255),
    winner_id VARCHAR(255),
    side_pots JSONB DEFAULT '[]',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tp_games_table ON teen_patti_games(table_id);
CREATE INDEX IF NOT EXISTS idx_tp_games_state ON teen_patti_games(state);

-- 2. Create teen_patti_players table
CREATE TABLE IF NOT EXISTS teen_patti_players (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id UUID REFERENCES teen_patti_games(id),
    user_id VARCHAR(255) NOT NULL,
    cards JSONB, -- Array of {suit, value}
    status VARCHAR(50) NOT NULL, -- ACTIVE, FOLDED, ALL_IN, LEFT
    is_blind BOOLEAN DEFAULT TRUE,
    seen_cards BOOLEAN DEFAULT FALSE,
    total_bet DECIMAL(12,2) DEFAULT 0,
    seat_index INT NOT NULL,
    hand_rank JSONB, -- {type: "TRAIL", values: [14, 14, 14]}
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tp_players_game ON teen_patti_players(game_id);
CREATE INDEX IF NOT EXISTS idx_tp_players_user ON teen_patti_players(user_id);

-- 3. Create view for game history
CREATE OR REPLACE VIEW teen_patti_history AS
SELECT
    g.id as game_id,
    g.table_id,
    g.pot_amount,
    g.winner_id,
    g.created_at,
    COUNT(p.id) as player_count,
    MAX(p.total_bet) as max_bet
FROM teen_patti_games g
JOIN teen_patti_players p ON g.id = p.game_id
WHERE g.state = 'FINISHED'
GROUP BY g.id, g.table_id, g.pot_amount, g.winner_id, g.created_at;

COMMENT ON TABLE teen_patti_games IS 'Active and historical Teen Patti game sessions';
COMMENT ON TABLE teen_patti_players IS 'Player participation and hands in Teen Patti games';
