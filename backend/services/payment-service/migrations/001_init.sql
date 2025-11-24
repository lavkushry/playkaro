-- Payment Service Database Schema
-- Database: payments_db

-- Payment orders table
CREATE TABLE IF NOT EXISTS payment_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    order_id VARCHAR(100) UNIQUE NOT NULL,
    gateway VARCHAR(50) NOT NULL,
    gateway_order_id VARCHAR(255),
    amount DECIMAL(15, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'INR',
    type VARCHAR(20) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'INITIATED',
    payment_method VARCHAR(50),
    metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_payment_orders_user_id ON payment_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_orders_status ON payment_orders(status);
CREATE INDEX IF NOT EXISTS idx_payment_orders_created_at ON payment_orders(created_at);

-- Payment transactions table
CREATE TABLE IF NOT EXISTS payment_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES payment_orders(id),
    gateway_txn_id VARCHAR(255),
    amount DECIMAL(15, 2) NOT NULL,
    fee DECIMAL(15, 2) DEFAULT 0,
    tax DECIMAL(15, 2) DEFAULT 0,
    net_amount DECIMAL(15, 2),
    reconciled BOOLEAN DEFAULT FALSE,
    reconciled_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Fraud detection logs
CREATE TABLE IF NOT EXISTS fraud_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    check_type VARCHAR(50) NOT NULL,
    risk_score INT,
    flagged BOOLEAN DEFAULT FALSE,
    details JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fraud_checks_user_id ON fraud_checks(user_id);
CREATE INDEX IF NOT EXISTS idx_fraud_checks_flagged ON fraud_checks(flagged);

-- Webhook logs
CREATE TABLE IF NOT EXISTS webhook_logs (
    id SERIAL PRIMARY KEY,
    gateway VARCHAR(50) NOT NULL,
    event_type VARCHAR(100),
    payload JSONB,
    signature VARCHAR(500),
    signature_valid BOOLEAN,
    processed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_webhook_logs_gateway ON webhook_logs(gateway);
CREATE INDEX IF NOT EXISTS idx_webhook_logs_processed ON webhook_logs(processed);
