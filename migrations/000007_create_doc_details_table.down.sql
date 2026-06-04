-- Migration rollback: drop doc_details table

DROP INDEX IF EXISTS idx_doc_details_status;
DROP INDEX IF EXISTS idx_doc_details_user_id;
DROP INDEX IF EXISTS idx_doc_details_doc_type_id;
DROP TABLE IF EXISTS doc_details CASCADE;
DROP TYPE IF EXISTS document_status;
