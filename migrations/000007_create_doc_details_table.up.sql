-- Migration: create doc_details table
-- doc_details depends on doc_types, users

CREATE TYPE document_status AS ENUM ('none', 'pending', 'waiting_approval', 'approved');

CREATE TABLE IF NOT EXISTS doc_details (
    id               UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    doc_no           VARCHAR(100)     NOT NULL UNIQUE,
    doc_name         VARCHAR(255)     NOT NULL,
    description      TEXT,
    version_number   INT              NOT NULL DEFAULT 1,
    status           document_status  NOT NULL DEFAULT 'none',
    doc_type_id      UUID             REFERENCES doc_types(id) ON DELETE SET NULL,
    user_id          UUID             REFERENCES users(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_doc_details_doc_type_id   ON doc_details(doc_type_id);
CREATE INDEX IF NOT EXISTS idx_doc_details_user_id       ON doc_details(user_id);
CREATE INDEX IF NOT EXISTS idx_doc_details_status        ON doc_details(status);

COMMENT ON TABLE doc_details IS 'Stores core metadata of electronic documents';
