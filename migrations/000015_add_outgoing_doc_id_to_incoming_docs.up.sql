-- incoming_docs must reference the specific outgoing_doc that created it,
-- not just doc_details_id. This prevents recipients from bleeding across
-- multiple uploads of the same file.

ALTER TABLE incoming_docs
ADD COLUMN outgoing_doc_id UUID REFERENCES outgoing_docs(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_incoming_docs_outgoing_doc_id ON incoming_docs(outgoing_doc_id);
