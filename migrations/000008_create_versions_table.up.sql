-- Migration: create versions table
-- versions depends on doc_details and folders

CREATE TABLE IF NOT EXISTS versions (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    doc_details_id UUID        NOT NULL REFERENCES doc_details(id) ON DELETE CASCADE,
    folder_id      UUID        REFERENCES folders(id) ON DELETE SET NULL,
    version_number INT         NOT NULL,
    doc_path       TEXT        NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ,
    UNIQUE (doc_details_id, version_number)
);

CREATE INDEX IF NOT EXISTS idx_versions_doc_details_id ON versions(doc_details_id);
CREATE INDEX IF NOT EXISTS idx_versions_folder_id ON versions(folder_id);

COMMENT ON TABLE versions IS 'Stores version history and file paths for documents';
