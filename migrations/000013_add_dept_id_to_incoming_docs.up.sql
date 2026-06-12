-- Migration: Add dept_id column to incoming_docs table
-- This tracks which department received the incoming document

ALTER TABLE incoming_docs
ADD COLUMN dept_id UUID REFERENCES departments(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_incoming_docs_dept_id ON incoming_docs(dept_id);

COMMENT ON COLUMN incoming_docs.dept_id IS 'Reference to the department that received this document';