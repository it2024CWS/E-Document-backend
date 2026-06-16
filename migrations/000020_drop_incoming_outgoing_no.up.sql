-- Migration: drop the per-record incoming_no / outgoing_no columns.
-- The shared doc_details.doc_no is now the single document identifier used
-- across the outgoing doc and every recipient's incoming record.

ALTER TABLE outgoing_docs DROP COLUMN IF EXISTS outgoing_no;
ALTER TABLE incoming_docs DROP COLUMN IF EXISTS incoming_no;

