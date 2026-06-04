-- Migration: create outgoing_docs table
-- outgoing_docs depends on doc_details, users, folders

CREATE TABLE IF NOT EXISTS outgoing_docs (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    outgoing_no    VARCHAR(100) NOT NULL UNIQUE,
    doc_details_id UUID        NOT NULL REFERENCES doc_details(id) ON DELETE CASCADE,
    folder_id      UUID        REFERENCES folders(id) ON DELETE SET NULL,
    created_by     UUID        REFERENCES users(id) ON DELETE SET NULL,
    updated_by     UUID        REFERENCES users(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_outgoing_docs_doc_details_id ON outgoing_docs(doc_details_id);
CREATE INDEX IF NOT EXISTS idx_outgoing_docs_folder_id      ON outgoing_docs(folder_id);
CREATE INDEX IF NOT EXISTS idx_outgoing_docs_created_by     ON outgoing_docs(created_by);
CREATE INDEX IF NOT EXISTS idx_outgoing_docs_updated_by     ON outgoing_docs(updated_by);

COMMENT ON TABLE outgoing_docs IS 'Tracks documents sent out of the organization or department';
