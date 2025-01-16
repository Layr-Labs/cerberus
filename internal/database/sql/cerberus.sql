CREATE SCHEMA IF NOT EXISTS public;

CREATE TABLE IF NOT EXISTS public.keys_metadata (
    public_key_g1 VARCHAR(255) PRIMARY KEY,
    public_key_g2 VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    api_key_hash text,
    locked boolean DEFAULT false
);
