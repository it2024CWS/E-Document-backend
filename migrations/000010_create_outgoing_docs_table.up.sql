-- Migration: create outgoing_docs table
-- outgoing_docs depends on docs, users, departments

CREATE TABLE IF NOT EXISTS outgoing_docs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    outgoing_no VARCHAR(100) NOT NULL UNIQUE,
    doc_id      UUID        NOT NULL REFERENCES docs(id) ON DELETE CASCADE,
    user_id     UUID        REFERENCES users(id) ON DELETE SET NULL,
    dept_id     UUID        REFERENCES departments(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_outgoing_docs_doc_id  ON outgoing_docs(doc_id);
CREATE INDEX IF NOT EXISTS idx_outgoing_docs_user_id ON outgoing_docs(user_id);
CREATE INDEX IF NOT EXISTS idx_outgoing_docs_dept_id ON outgoing_docs(dept_id);

COMMENT ON TABLE outgoing_docs IS 'Tracks documents sent out of the organization or department';
