-- Migration: create folders table
-- folders depends on users; supports self-referential parent_folder_id

CREATE TABLE IF NOT EXISTS folders (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    folder_name      VARCHAR(255) NOT NULL,
    folder_path      TEXT        NOT NULL,
    user_id          UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_folder_id UUID        REFERENCES folders(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_folders_user_id          ON folders(user_id);
CREATE INDEX IF NOT EXISTS idx_folders_parent_folder_id ON folders(parent_folder_id);

COMMENT ON TABLE folders IS 'Stores folder hierarchy for organizing documents per user';
