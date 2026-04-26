-- Add folder_id to versions table
ALTER TABLE versions
    ADD COLUMN IF NOT EXISTS folder_id UUID REFERENCES folders(id) ON DELETE SET NULL;

-- Index for performance
CREATE INDEX IF NOT EXISTS idx_versions_folder_id ON versions(folder_id);
