-- incoming_no should only be assigned when the Secretary receives the document,
-- not at outgoing_doc creation time. NULL values do not conflict with UNIQUE.

ALTER TABLE incoming_docs
    ALTER COLUMN incoming_no DROP NOT NULL,
    ALTER COLUMN incoming_no SET DEFAULT NULL;
