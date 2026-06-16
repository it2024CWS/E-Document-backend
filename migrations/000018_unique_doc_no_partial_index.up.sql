-- Migration: make doc_no uniqueness ignore soft-deleted documents.
-- The original column-level UNIQUE constraint (doc_details_doc_no_key) is global,
-- so a soft-deleted document's number could never be reused — and worse, the
-- service-level availability check (which filters deleted_at IS NULL) would report
-- such a number as "available" only for the insert to then fail on the constraint.
-- Replace it with a partial unique index that only applies to live rows.

ALTER TABLE doc_details DROP CONSTRAINT IF EXISTS doc_details_doc_no_key;

CREATE UNIQUE INDEX IF NOT EXISTS uq_doc_details_doc_no_active
    ON doc_details (doc_no)
    WHERE deleted_at IS NULL;
