DROP INDEX IF EXISTS idx_outgoing_docs_status;

ALTER TABLE outgoing_docs DROP COLUMN IF EXISTS status;
