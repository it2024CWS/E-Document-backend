-- Migration: create docs table
-- docs depends on doc_types, folders, users

CREATE TYPE document_status AS ENUM ('none', 'pending', 'waiting_approval', 'approved');

CREATE TABLE IF NOT EXISTS docs (
    id               UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    doc_no           VARCHAR(100)     NOT NULL UNIQUE,
    doc_name         VARCHAR(255)     NOT NULL,
    doc_path         TEXT             NOT NULL DEFAULT '',
    type             VARCHAR(20)      NOT NULL DEFAULT '',  -- file extension: pdf, doc, excel
    doc_type_id      UUID             REFERENCES doc_types(id) ON DELETE SET NULL,
    folder_id        UUID             REFERENCES folders(id) ON DELETE SET NULL,
    registrant_id    UUID             REFERENCES users(id) ON DELETE SET NULL,
    status           document_status  NOT NULL DEFAULT 'none',
    version_number   INT              NOT NULL DEFAULT 1,
    description      TEXT,
    send_to_director BOOLEAN          NOT NULL DEFAULT FALSE,
    created_at       TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_docs_doc_type_id   ON docs(doc_type_id);
CREATE INDEX IF NOT EXISTS idx_docs_folder_id     ON docs(folder_id);
CREATE INDEX IF NOT EXISTS idx_docs_registrant_id ON docs(registrant_id);
CREATE INDEX IF NOT EXISTS idx_docs_status        ON docs(status);

COMMENT ON TABLE docs IS 'Stores all electronic documents in the system';
