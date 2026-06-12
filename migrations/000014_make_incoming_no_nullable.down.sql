-- Revert: restore NOT NULL on incoming_no (requires all existing rows to have a value)
UPDATE incoming_docs SET incoming_no = gen_random_uuid()::text WHERE incoming_no IS NULL;

ALTER TABLE incoming_docs
    ALTER COLUMN incoming_no SET NOT NULL,
    ALTER COLUMN incoming_no DROP DEFAULT;
