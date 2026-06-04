-- Migration: create sectors table
-- sectors depends on departments

CREATE TABLE IF NOT EXISTS sectors (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(150) NOT NULL,
    dept_id     UUID        NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (name, dept_id)
);

CREATE INDEX IF NOT EXISTS idx_sectors_dept_id ON sectors(dept_id);

COMMENT ON TABLE sectors IS 'Stores sector (sub-unit) information within a department';
