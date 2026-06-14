-- Add file_type column to versions table to store MIME type or file extension
ALTER TABLE versions
ADD COLUMN file_type VARCHAR(50);

-- Create index for file_type queries if needed
CREATE INDEX idx_versions_file_type ON versions(file_type);
