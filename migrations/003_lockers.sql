-- Migration: 003_lockers.sql
-- Description: Physical lockers and their individual slots

CREATE TABLE IF NOT EXISTS lockers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    device_id VARCHAR(100) NOT NULL UNIQUE,
    status VARCHAR(30) NOT NULL DEFAULT 'online' CHECK (status IN ('online', 'offline', 'maintenance')),
    firmware_version VARCHAR(50) DEFAULT '1.0.0',
    last_seen_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS locker_slots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    locker_id UUID NOT NULL REFERENCES lockers(id) ON DELETE CASCADE,
    slot_code VARCHAR(20) NOT NULL, -- e.g. A1, A2, B1, E5
    size VARCHAR(20) NOT NULL DEFAULT 'Medium' CHECK (size IN ('Small', 'Medium', 'Large', 'ExtraLarge')),
    status VARCHAR(30) NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'reserved', 'occupied', 'maintenance', 'offline')),
    base_price_per_hour NUMERIC(12, 2) NOT NULL DEFAULT 5000.00,
    current_rental_id UUID, -- References active rental if any
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT uq_locker_slot_code UNIQUE(locker_id, slot_code)
);

CREATE INDEX IF NOT EXISTS idx_lockers_location ON lockers(location_id);
CREATE INDEX IF NOT EXISTS idx_lockers_device ON lockers(device_id);
CREATE INDEX IF NOT EXISTS idx_slots_locker_status ON locker_slots(locker_id, status);
