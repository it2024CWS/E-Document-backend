-- Create folders table for hierarchical folder structure
CREATE TABLE folders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    path TEXT NOT NULL,
    is_root_folder BOOLEAN NOT NULL DEFAULT false,
    parent_folder_id UUID REFERENCES folders(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_folders_parent ON folders(parent_folder_id);
CREATE INDEX idx_folders_owner ON folders(owner_id);
CREATE INDEX idx_folders_path ON folders(path);

-- Unique constraint: folder name must be unique within same parent and owner
CREATE UNIQUE INDEX idx_folders_unique_name ON folders(name, COALESCE(parent_folder_id, '00000000-0000-0000-0000-000000000000'::uuid), owner_id);
