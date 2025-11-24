-- Migration 018: Live Streaming
-- Implements stream key management and live status tracking

CREATE TABLE IF NOT EXISTS streams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id VARCHAR(255) NOT NULL,
    stream_key VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'OFFLINE',
    viewer_count INTEGER NOT NULL DEFAULT 0,
    playback_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_streams_match ON streams(match_id);
CREATE INDEX IF NOT EXISTS idx_streams_key ON streams(stream_key);
CREATE INDEX IF NOT EXISTS idx_streams_status ON streams(status);

COMMENT ON TABLE streams IS 'Manages live stream keys and status for matches/games';
COMMENT ON COLUMN streams.stream_key IS 'Secret key for RTMP publishing';
COMMENT ON COLUMN streams.playback_url IS 'HLS playback URL for viewers';
