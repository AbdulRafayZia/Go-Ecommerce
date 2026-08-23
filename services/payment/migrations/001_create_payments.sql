-- Create payments table
CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY,
    order_id UUID NOT NULL,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'authorized', 'captured', 'failed', 'cancelled', 'refunded')),
    payment_method VARCHAR(20) NOT NULL CHECK (payment_method IN ('card', 'bank_transfer', 'wallet')),
    provider_payment_id VARCHAR(255),
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    failure_reason TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_payments_order_id ON payments(order_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_provider_payment_id ON payments(provider_payment_id) WHERE provider_payment_id IS NOT NULL;
CREATE INDEX idx_payments_created_at ON payments(created_at DESC);
CREATE INDEX idx_payments_idempotency_key ON payments(idempotency_key);

-- Create composite index for common queries
CREATE INDEX idx_payments_order_status ON payments(order_id, status);

-- Comments
COMMENT ON TABLE payments IS 'Stores payment records for orders';
COMMENT ON COLUMN payments.id IS 'Unique payment identifier';
COMMENT ON COLUMN payments.order_id IS 'Reference to the order this payment is for';
COMMENT ON COLUMN payments.amount IS 'Payment amount in the specified currency';
COMMENT ON COLUMN payments.currency IS 'ISO 4217 currency code (e.g., USD, EUR)';
COMMENT ON COLUMN payments.status IS 'Current payment status';
COMMENT ON COLUMN payments.payment_method IS 'Payment method type';
COMMENT ON COLUMN payments.provider_payment_id IS 'Payment ID from payment provider (e.g., Stripe charge ID)';
COMMENT ON COLUMN payments.idempotency_key IS 'Key to ensure idempotent payment creation';
COMMENT ON COLUMN payments.failure_reason IS 'Reason for payment failure (if applicable)';
COMMENT ON COLUMN payments.metadata IS 'Additional payment metadata in JSON format';
COMMENT ON COLUMN payments.created_at IS 'Timestamp when payment was created';
COMMENT ON COLUMN payments.updated_at IS 'Timestamp when payment was last updated';
