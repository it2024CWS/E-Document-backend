-- Increase length of type column in docs table to hold longer MIME types
ALTER TABLE docs ALTER COLUMN type TYPE VARCHAR(255);
