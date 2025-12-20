-- Create document type enum
CREATE TYPE document_type AS ENUM ('General', 'Barcode');

-- Create document status enum
CREATE TYPE document_status AS ENUM ('Draft', 'Pending', 'Approved', 'Rejected');

-- Create documents table
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    type document_type NOT NULL DEFAULT 'General',
    category_id UUID,
    folder_id UUID REFERENCES folders(id) ON DELETE SET NULL,
    barcode VARCHAR(100) UNIQUE,
    registrant_id UUID REFERENCES users(id),
    current_department_id UUID,
    status document_status DEFAULT 'Draft',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_documents_folder ON documents(folder_id);
CREATE INDEX idx_documents_barcode ON documents(barcode) WHERE barcode IS NOT NULL;
CREATE INDEX idx_documents_registrant ON documents(registrant_id);
CREATE INDEX idx_documents_status ON documents(status);
