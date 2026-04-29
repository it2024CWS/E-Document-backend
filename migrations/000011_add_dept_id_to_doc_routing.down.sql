DROP INDEX IF EXISTS idx_outgoing_docs_dept_id;
DROP INDEX IF EXISTS idx_incoming_docs_dept_id;
ALTER TABLE outgoing_docs DROP COLUMN IF EXISTS dept_id;
ALTER TABLE incoming_docs DROP COLUMN IF EXISTS dept_id;
