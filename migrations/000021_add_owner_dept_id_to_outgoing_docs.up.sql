ALTER TABLE outgoing_docs
    ADD COLUMN IF NOT EXISTS owner_dept_id UUID REFERENCES departments(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_outgoing_docs_owner_dept_id ON outgoing_docs(owner_dept_id);
