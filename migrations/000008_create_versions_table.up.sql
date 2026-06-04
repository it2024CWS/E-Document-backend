-- Migration: create versions table
-- versions depends on docs and folders

CREATE TABLE IF NOT EXISTS versions (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    doc_id         UUID        NOT NULL REFERENCES docs(id) ON DELETE CASCADE,
    folder_id      UUID        REFERENCES folders(id) ON DELETE SET NULL,
    version_number INT         NOT NULL,
    doc_path       TEXT        NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (doc_id, version_number)
);

CREATE INDEX IF NOT EXISTS idx_versions_doc_id ON versions(doc_id);

COMMENT ON TABLE versions IS 'Stores version history for each document';
