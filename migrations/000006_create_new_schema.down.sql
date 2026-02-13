-- Rollback: Drop all new tables in reverse order
DROP TRIGGER IF EXISTS update_docs_updated_at ON docs;
DROP TRIGGER IF EXISTS update_doc_types_updated_at ON doc_types;
DROP TRIGGER IF EXISTS update_sectors_updated_at ON sectors;
DROP TRIGGER IF EXISTS update_departments_updated_at ON departments;
DROP TRIGGER IF EXISTS update_user_roles_updated_at ON user_roles;

DROP TABLE IF EXISTS versions CASCADE;
DROP TABLE IF EXISTS outgoing_docs CASCADE;
DROP TABLE IF EXISTS incoming_docs CASCADE;
DROP TABLE IF EXISTS docs CASCADE;
DROP TABLE IF EXISTS folders CASCADE;
DROP TABLE IF EXISTS doc_types CASCADE;
DROP TABLE IF EXISTS sectors CASCADE;
DROP TABLE IF EXISTS departments CASCADE;
DROP TABLE IF EXISTS user_roles CASCADE;

-- Restore users table to old structure
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_sector;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_department;

ALTER TABLE users ADD COLUMN IF NOT EXISTS first_name VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_name VARCHAR(255);

UPDATE users SET first_name = firstname WHERE first_name IS NULL;
UPDATE users SET last_name = lastname WHERE last_name IS NULL;

ALTER TABLE users DROP COLUMN IF EXISTS firstname;
ALTER TABLE users DROP COLUMN IF EXISTS lastname;
ALTER TABLE users DROP COLUMN IF EXISTS nickname;
ALTER TABLE users DROP COLUMN IF EXISTS is_active;
ALTER TABLE users DROP COLUMN IF EXISTS role_id;

ALTER TABLE users ALTER COLUMN department_id TYPE VARCHAR(255);
ALTER TABLE users ALTER COLUMN sector_id TYPE VARCHAR(255);

ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50) CHECK (role IN ('Director', 'DepartmentManager', 'SectorManager', 'Employee'));
