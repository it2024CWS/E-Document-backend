DROP INDEX IF EXISTS idx_incoming_docs_outgoing_doc_id;

ALTER TABLE incoming_docs
DROP COLUMN outgoing_doc_id;
