-- Migration: add a status gate to outgoing_docs.
-- The owner department head now approves the document on the OUTGOING side
-- (instead of as the first incoming-doc step). This status tracks that gate:
--   pending   — awaiting the owner department head's approval (initial state)
--   approved  — owner head approved; the sequential incoming-doc flow has started
--   rejected  — owner head rejected; no recipient incoming docs are created

ALTER TABLE outgoing_docs
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'pending';

CREATE INDEX IF NOT EXISTS idx_outgoing_docs_status ON outgoing_docs(status);
