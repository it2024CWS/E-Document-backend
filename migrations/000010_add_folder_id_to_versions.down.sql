-- Revert: remove folder_id from versions table
DROP INDEX IF EXISTS idx_versions_folder_id;
ALTER TABLE versions DROP COLUMN IF EXISTS folder_id;
