-- Migration: 001_profiles.sql
-- Description: User profiles connected to Supabase Auth (auth.users)

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS profiles (
    id UUID PRIMARY KEY, -- Can reference auth.users(id) in Supabase
    email VARCHAR(255) NOT NULL UNIQUE,
    username VARCHAR(100) UNIQUE,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(30),
    avatar_url TEXT,
    role VARCHAR(30) NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin', 'operator', 'viewer')),
    status VARCHAR(30) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deactivated')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_profiles_email ON profiles(email);
CREATE INDEX IF NOT EXISTS idx_profiles_role ON profiles(role);
