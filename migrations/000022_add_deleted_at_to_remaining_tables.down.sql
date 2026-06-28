ALTER TABLE user_roles    DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE departments   DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE sectors       DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE users         DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE doc_types     DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE outgoing_docs DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE incoming_docs DROP COLUMN IF EXISTS deleted_at;
