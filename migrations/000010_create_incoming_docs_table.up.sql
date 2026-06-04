-- Migration: create incoming_docs table
-- incoming_docs depends on doc_details, users, folders

CREATE TYPE incoming_doc_status AS ENUM ('pending', 'received', 'approved', 'rejected');

CREATE TABLE IF NOT EXISTS incoming_docs (
    id             UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    incoming_no    VARCHAR(100)        NOT NULL UNIQUE,
    incoming_date  TIMESTAMPTZ,
    received_date  TIMESTAMPTZ,
    status         incoming_doc_status NOT NULL DEFAULT 'pending',
    doc_details_id UUID                NOT NULL REFERENCES doc_details(id) ON DELETE CASCADE,
    folder_id      UUID                REFERENCES folders(id) ON DELETE SET NULL,
    created_by     UUID                REFERENCES users(id) ON DELETE SET NULL,
    updated_by     UUID                REFERENCES users(id) ON DELETE SET NULL,
    approver_id    UUID                REFERENCES users(id) ON DELETE SET NULL,
    approver_date  TIMESTAMPTZ,
    remark         TEXT                NOT NULL DEFAULT '',
    updated_at     TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_incoming_docs_doc_details_id ON incoming_docs(doc_details_id);
CREATE INDEX IF NOT EXISTS idx_incoming_docs_folder_id      ON incoming_docs(folder_id);
CREATE INDEX IF NOT EXISTS idx_incoming_docs_created_by     ON incoming_docs(created_by);
CREATE INDEX IF NOT EXISTS idx_incoming_docs_updated_by     ON incoming_docs(updated_by);
CREATE INDEX IF NOT EXISTS idx_incoming_docs_approver_id    ON incoming_docs(approver_id);
CREATE INDEX IF NOT EXISTS idx_incoming_docs_status         ON incoming_docs(status);

COMMENT ON TABLE incoming_docs IS 'Tracks documents received by departments/users, including approval workflow';
