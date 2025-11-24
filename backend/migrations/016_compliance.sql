-- Migration 016: Compliance & Security
-- Stores KYC requests and responsible gaming limits

-- 1. Create kyc_requests table
CREATE TABLE IF NOT EXISTS kyc_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    document_type VARCHAR(50) NOT NULL, -- AADHAAR, PAN
    document_number TEXT NOT NULL, -- Encrypted
    image_url TEXT NOT NULL, -- Encrypted
    status VARCHAR(50) NOT NULL, -- PENDING, VERIFIED, REJECTED
    admin_notes TEXT,
    verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_kyc_user ON kyc_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_kyc_status ON kyc_requests(status);

-- 2. Create user_limits table
CREATE TABLE IF NOT EXISTS user_limits (
    user_id VARCHAR(255) PRIMARY KEY,
    deposit_daily DECIMAL(12,2) DEFAULT 0, -- 0 means no limit
    deposit_weekly DECIMAL(12,2) DEFAULT 0,
    deposit_monthly DECIMAL(12,2) DEFAULT 0,
    self_exclusion_end TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE kyc_requests IS 'User identity verification requests (Encrypted PII)';
COMMENT ON TABLE user_limits IS 'Responsible gaming limits and self-exclusion';
