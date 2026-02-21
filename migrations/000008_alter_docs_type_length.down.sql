-- Revert length of type column in docs table back to 50
ALTER TABLE docs ALTER COLUMN type TYPE VARCHAR(50);
