-- Migration: 004_rentals.sql
-- Description: Rental lifecycle, price snapshot, and slot concurrency protection

CREATE TABLE IF NOT EXISTS rentals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES profiles(id) ON DELETE RESTRICT,
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE RESTRICT,
    locker_id UUID NOT NULL REFERENCES lockers(id) ON DELETE RESTRICT,
    slot_id UUID NOT NULL REFERENCES locker_slots(id) ON DELETE RESTRICT,
    status VARCHAR(30) NOT NULL DEFAULT 'draft' CHECK (
        status IN ('draft', 'reserved', 'awaiting_payment', 'paid', 'active', 'completed', 'expired', 'cancelled')
    ),
    duration_hours INT NOT NULL CHECK (duration_hours > 0),
    base_price_snapshot NUMERIC(12, 2) NOT NULL,
    discount_amount NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    total_amount NUMERIC(12, 2) NOT NULL,
    reservation_expires_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Partial unique index to enforce strict concurrency protection:
-- A locker slot can have at most ONE active rental (reserved, awaiting_payment, paid, active)
CREATE UNIQUE INDEX IF NOT EXISTS uq_active_slot_rental 
ON rentals (slot_id) 
WHERE status IN ('reserved', 'awaiting_payment', 'paid', 'active');

CREATE INDEX IF NOT EXISTS idx_rentals_user ON rentals(user_id);
CREATE INDEX IF NOT EXISTS idx_rentals_status ON rentals(status);
CREATE INDEX IF NOT EXISTS idx_rentals_res_expiry ON rentals(reservation_expires_at) WHERE status = 'reserved';
