-- Migration 014: Social Features
-- Stores friendships, chat history, and user profiles

-- 1. Create friendships table
CREATE TABLE IF NOT EXISTS friendships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    requester_id VARCHAR(255) NOT NULL,
    addressee_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL, -- PENDING, ACCEPTED, BLOCKED
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(requester_id, addressee_id)
);

CREATE INDEX IF NOT EXISTS idx_friendships_user ON friendships(requester_id);
CREATE INDEX IF NOT EXISTS idx_friendships_addressee ON friendships(addressee_id);
CREATE INDEX IF NOT EXISTS idx_friendships_status ON friendships(status);

-- 2. Create user_profiles table (Enhanced)
CREATE TABLE IF NOT EXISTS user_profiles (
    user_id VARCHAR(255) PRIMARY KEY,
    bio TEXT,
    avatar_frame VARCHAR(255),
    title VARCHAR(255),
    stats JSONB DEFAULT '{}', -- {wins: 10, rank: "Gold"}
    last_active TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Create chat_messages table (Archive)
CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id VARCHAR(255) NOT NULL,
    recipient_id VARCHAR(255), -- NULL for global
    content TEXT NOT NULL,
    type VARCHAR(50) NOT NULL, -- GLOBAL, PRIVATE, SYSTEM
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_chat_sender ON chat_messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_chat_recipient ON chat_messages(recipient_id);
CREATE INDEX IF NOT EXISTS idx_chat_created ON chat_messages(created_at DESC);

COMMENT ON TABLE friendships IS 'User relationships (friends, blocks)';
COMMENT ON TABLE user_profiles IS 'Enhanced user data for social display';
COMMENT ON TABLE chat_messages IS 'Archived chat history';
