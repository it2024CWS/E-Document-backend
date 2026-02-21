-- Add registrant_id column to docs table (the user who created/owns the document)
ALTER TABLE docs ADD COLUMN IF NOT EXISTS registrant_id UUID REFERENCES users(id) ON DELETE SET NULL;

-- Index for performance
CREATE INDEX IF NOT EXISTS idx_docs_registrant_id ON docs(registrant_id);
