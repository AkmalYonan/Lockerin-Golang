-- Migration: 006_payments.sql
-- Description: Payment logs, provider transaction references, and webhook idempotency

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rental_id UUID NOT NULL REFERENCES rentals(id) ON DELETE RESTRICT,
    provider VARCHAR(50) NOT NULL DEFAULT 'midtrans',
    order_id VARCHAR(100) NOT NULL UNIQUE,
    provider_transaction_id VARCHAR(100),
    amount NUMERIC(12, 2) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending' CHECK (
        status IN ('pending', 'settlement', 'capture', 'deny', 'cancel', 'expire', 'failure', 'refund')
    ),
    payment_type VARCHAR(50), -- e.g. qris, gopay, bank_transfer, cstore
    snap_token TEXT,
    snap_redirect_url TEXT,
    raw_response JSONB,
    paid_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_rental ON payments(rental_id);
CREATE INDEX IF NOT EXISTS idx_payments_order ON payments(order_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
