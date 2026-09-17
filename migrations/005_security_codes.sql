-- Migration: 005_security_codes.sql
-- Description: One-time PINs stored strictly as hashes with rate limiting and TTL

CREATE TABLE IF NOT EXISTS security_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rental_id UUID NOT NULL UNIQUE REFERENCES rentals(id) ON DELETE CASCADE,
    pin_hash VARCHAR(255) NOT NULL,
    salt VARCHAR(64) NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 5,
    locked_until TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_security_codes_rental ON security_codes(rental_id);
CREATE INDEX IF NOT EXISTS idx_security_codes_expiry ON security_codes(expires_at);
