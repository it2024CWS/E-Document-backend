-- Migration: add soft-delete column to the remaining business tables.
-- Brings them in line with folders / doc_details / versions, which already
-- carry deleted_at and are filtered by `WHERE deleted_at IS NULL` in their
-- repository queries. outgoing_doc_routes is intentionally skipped — it is a
-- pure routing table with ON DELETE CASCADE on both its parent FKs.

ALTER TABLE user_roles    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE departments   ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE sectors       ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE users         ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE doc_types     ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE outgoing_docs ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE incoming_docs ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
