-- Migration: 010_pricing_tiers.sql
-- Description: Pricing tiers master table

CREATE TABLE IF NOT EXISTS pricing_tiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slot_size VARCHAR(20) NOT NULL UNIQUE,
    hourly_rate NUMERIC(12, 2) NOT NULL,
    deposit_amount NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Seed initial pricing tiers if not present
INSERT INTO pricing_tiers (slot_size, hourly_rate, deposit_amount, updated_at)
VALUES 
    ('small', 4000.00, 10000.00, NOW()),
    ('medium', 6000.00, 15000.00, NOW()),
    ('large', 9000.00, 20000.00, NOW())
ON CONFLICT (slot_size) DO UPDATE 
SET hourly_rate = EXCLUDED.hourly_rate, deposit_amount = EXCLUDED.deposit_amount, updated_at = NOW();
