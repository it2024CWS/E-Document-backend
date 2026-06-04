-- Rollback: drop incoming_docs table and incoming_doc_status type
DROP TABLE IF EXISTS incoming_docs;
DROP TYPE IF EXISTS incoming_doc_status;
