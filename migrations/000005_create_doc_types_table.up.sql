-- Migration: create doc_types table
-- doc_types has no FK dependencies

CREATE TABLE IF NOT EXISTS doc_types (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    type_name   VARCHAR(150) NOT NULL UNIQUE,
    description TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE doc_types IS 'Stores document type categories (e.g. memo, letter, report)';
