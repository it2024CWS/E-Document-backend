-- Create user_roles table
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert default roles
INSERT INTO user_roles (role_name, description) VALUES
    ('Employee', 'ພະນັກງານທົ່ວໄປ'),
    ('DepartmentSecretary', 'ເລຂາພະແນກ'),
    ('DepartmentHead', 'ຫົວໜ້າພະແນກ'),
    ('DirectorSecretary', 'ເລຂາຜູ້ອຳນວຍການ'),
    ('Director', 'ຜູ້ອຳນວຍການ');

-- Create departments table
CREATE TABLE IF NOT EXISTS departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dept_name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create sectors table
CREATE TABLE IF NOT EXISTS sectors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    dept_id UUID REFERENCES departments(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Update users table structure
ALTER TABLE users DROP COLUMN IF EXISTS role;
ALTER TABLE users ADD COLUMN IF NOT EXISTS role_id UUID REFERENCES user_roles(id);
ALTER TABLE users ADD COLUMN IF NOT EXISTS firstname VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS lastname VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS nickname VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- Update existing data: migrate first_name to firstname, last_name to lastname
UPDATE users SET firstname = first_name WHERE firstname IS NULL;
UPDATE users SET lastname = last_name WHERE lastname IS NULL;

-- Drop old columns
ALTER TABLE users DROP COLUMN IF EXISTS first_name;
ALTER TABLE users DROP COLUMN IF EXISTS last_name;

-- Update department_id and sector_id to be UUIDs (they were VARCHAR before)
ALTER TABLE users ALTER COLUMN department_id TYPE UUID USING NULLIF(department_id, '')::uuid;
ALTER TABLE users ALTER COLUMN sector_id TYPE UUID USING NULLIF(sector_id, '')::uuid;

-- Add foreign keys
ALTER TABLE users ADD CONSTRAINT fk_users_department 
    FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE SET NULL;
ALTER TABLE users ADD CONSTRAINT fk_users_sector 
    FOREIGN KEY (sector_id) REFERENCES sectors(id) ON DELETE SET NULL;

-- Create doc_types table
CREATE TABLE IF NOT EXISTS doc_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type_name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create folders table with self-reference
CREATE TABLE IF NOT EXISTS folders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    folder_name VARCHAR(255) NOT NULL,
    folder_path TEXT NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    parent_folder_id UUID REFERENCES folders(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create docs table
CREATE TABLE IF NOT EXISTS docs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doc_no VARCHAR(50) UNIQUE NOT NULL,
    doc_name VARCHAR(255) NOT NULL,
    doc_path TEXT,
    type VARCHAR(50),
    doc_type_id UUID REFERENCES doc_types(id) ON DELETE SET NULL,
    folder_id UUID REFERENCES folders(id) ON DELETE SET NULL,
    status VARCHAR(50) DEFAULT 'none' CHECK (status IN ('none', 'pending', 'waiting_approval', 'approved')),
    version_number INTEGER DEFAULT 1,
    description TEXT,
    send_to_director BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create incoming_docs table
CREATE TABLE IF NOT EXISTS incoming_docs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incoming_no VARCHAR(50) UNIQUE NOT NULL,
    doc_id UUID REFERENCES docs(id) ON DELETE CASCADE,
    sender_id UUID REFERENCES users(id) ON DELETE SET NULL,
    receiver_id UUID REFERENCES users(id) ON DELETE SET NULL,
    approver_id UUID REFERENCES users(id) ON DELETE SET NULL,
    received_date TIMESTAMP WITH TIME ZONE,
    approver_date TIMESTAMP WITH TIME ZONE,
    remark TEXT,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'received', 'approved', 'rejected')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create outgoing_docs table
CREATE TABLE IF NOT EXISTS outgoing_docs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    outgoing_no VARCHAR(50) UNIQUE NOT NULL,
    doc_id UUID REFERENCES docs(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create versions table
CREATE TABLE IF NOT EXISTS versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    doc_id UUID REFERENCES docs(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    doc_path TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_sectors_dept_id ON sectors(dept_id);
CREATE INDEX idx_folders_user_id ON folders(user_id);
CREATE INDEX idx_folders_parent_id ON folders(parent_folder_id);
CREATE INDEX idx_docs_doc_type_id ON docs(doc_type_id);
CREATE INDEX idx_docs_folder_id ON docs(folder_id);
CREATE INDEX idx_docs_status ON docs(status);
CREATE INDEX idx_incoming_docs_doc_id ON incoming_docs(doc_id);
CREATE INDEX idx_incoming_docs_status ON incoming_docs(status);
CREATE INDEX idx_outgoing_docs_doc_id ON outgoing_docs(doc_id);
CREATE INDEX idx_versions_doc_id ON versions(doc_id);
CREATE INDEX idx_users_role_id ON users(role_id);

-- Create triggers for updated_at
CREATE TRIGGER update_user_roles_updated_at
    BEFORE UPDATE ON user_roles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_departments_updated_at
    BEFORE UPDATE ON departments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sectors_updated_at
    BEFORE UPDATE ON sectors
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_doc_types_updated_at
    BEFORE UPDATE ON doc_types
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_docs_updated_at
    BEFORE UPDATE ON docs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
