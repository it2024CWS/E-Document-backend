-- Revert: remove registrant_id from docs table
ALTER TABLE docs DROP COLUMN IF EXISTS registrant_id;
DROP INDEX IF EXISTS idx_docs_registrant_id;
