-- Migration rollback: drop versions table

DROP INDEX IF EXISTS idx_versions_folder_id;
DROP INDEX IF EXISTS idx_versions_doc_details_id;
DROP TABLE IF EXISTS versions CASCADE;
