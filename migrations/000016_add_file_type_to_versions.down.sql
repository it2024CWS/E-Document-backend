DROP INDEX IF EXISTS idx_versions_file_type;

ALTER TABLE versions
DROP COLUMN file_type;
