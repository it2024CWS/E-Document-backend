-- Revert to the global column-level UNIQUE constraint on doc_no.

DROP INDEX IF EXISTS uq_doc_details_doc_no_active;

ALTER TABLE doc_details ADD CONSTRAINT doc_details_doc_no_key UNIQUE (doc_no);
