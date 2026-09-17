-- Migration: 007_devices.sql
-- Description: IoT Head Locker command queue, execution tracking, and ACK verification

CREATE TABLE IF NOT EXISTS device_commands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    locker_id UUID NOT NULL REFERENCES lockers(id) ON DELETE CASCADE,
    slot_id UUID REFERENCES locker_slots(id) ON DELETE CASCADE,
    command VARCHAR(50) NOT NULL CHECK (command IN ('OPEN_SLOT', 'LOCK_SLOT', 'REBOOT', 'STATUS_CHECK')),
    correlation_id VARCHAR(100) NOT NULL UNIQUE,
    payload JSONB,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SENT', 'ACKNOWLEDGED', 'FAILED', 'TIMEOUT')),
    requested_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_device_commands_locker ON device_commands(locker_id);
CREATE INDEX IF NOT EXISTS idx_device_commands_correlation ON device_commands(correlation_id);
CREATE INDEX IF NOT EXISTS idx_device_commands_status ON device_commands(status);
