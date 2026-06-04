-- Migration: create document_attachments table
-- document_attachments depends on docs and users

CREATE TABLE IF NOT EXISTS document_attachments (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID        NOT NULL REFERENCES docs(id) ON DELETE CASCADE,
    file_name   VARCHAR(255) NOT NULL,
    file_path   TEXT        NOT NULL,
    file_size   BIGINT      NOT NULL DEFAULT 0,
    file_type   VARCHAR(50) NOT NULL DEFAULT '',
    version     INT         NOT NULL DEFAULT 1,
    is_current  BOOLEAN     NOT NULL DEFAULT TRUE,
    uploaded_by UUID        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_doc_attachments_document_id ON document_attachments(document_id);
CREATE INDEX IF NOT EXISTS idx_doc_attachments_uploaded_by ON document_attachments(uploaded_by);

COMMENT ON TABLE document_attachments IS 'Stores physical file attachments (from tusd upload pipeline) linked to documents';
