-- Migration: create departments table
-- departments has no FK dependencies

CREATE TABLE IF NOT EXISTS departments (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    dept_name   VARCHAR(150) NOT NULL UNIQUE,
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE departments IS 'Stores department information';
