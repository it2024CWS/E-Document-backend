-- Migration rollback: drop incoming_docs table

DROP INDEX IF EXISTS idx_incoming_docs_status;
DROP INDEX IF EXISTS idx_incoming_docs_approver_id;
DROP INDEX IF EXISTS idx_incoming_docs_updated_by;
DROP INDEX IF EXISTS idx_incoming_docs_created_by;
DROP INDEX IF EXISTS idx_incoming_docs_folder_id;
DROP INDEX IF EXISTS idx_incoming_docs_doc_details_id;
DROP TABLE IF EXISTS incoming_docs CASCADE;
DROP TYPE IF EXISTS incoming_doc_status;
