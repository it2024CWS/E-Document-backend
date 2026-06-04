-- Migration rollback: drop outgoing_docs table

DROP INDEX IF EXISTS idx_outgoing_docs_updated_by;
DROP INDEX IF EXISTS idx_outgoing_docs_created_by;
DROP INDEX IF EXISTS idx_outgoing_docs_folder_id;
DROP INDEX IF EXISTS idx_outgoing_docs_doc_details_id;
DROP TABLE IF EXISTS outgoing_docs CASCADE;
