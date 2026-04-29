-- Add dept_id to incoming_docs and outgoing_docs
ALTER TABLE incoming_docs ADD COLUMN IF NOT EXISTS dept_id UUID REFERENCES departments(id) ON DELETE SET NULL;
ALTER TABLE outgoing_docs ADD COLUMN IF NOT EXISTS dept_id UUID REFERENCES departments(id) ON DELETE SET NULL;

-- Index for performance
CREATE INDEX IF NOT EXISTS idx_incoming_docs_dept_id ON incoming_docs(dept_id);
CREATE INDEX IF NOT EXISTS idx_outgoing_docs_dept_id ON outgoing_docs(dept_id);
