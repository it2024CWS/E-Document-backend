-- Rollback: drop docs table and document_status type
DROP TABLE IF EXISTS docs;
DROP TYPE IF EXISTS document_status;
