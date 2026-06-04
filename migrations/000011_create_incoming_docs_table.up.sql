-- Migration: create incoming_docs table
-- incoming_docs depends on docs, users, departments

CREATE TYPE incoming_doc_status AS ENUM ('pending', 'received', 'approved', 'rejected');

CREATE TABLE IF NOT EXISTS incoming_docs (
    id            UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    incoming_no   VARCHAR(100)        NOT NULL UNIQUE,
    doc_id        UUID                NOT NULL REFERENCES docs(id) ON DELETE CASCADE,
    sender_id     UUID                REFERENCES users(id) ON DELETE SET NULL,
    receiver_id   UUID                REFERENCES users(id) ON DELETE SET NULL,
    approver_id   UUID                REFERENCES users(id) ON DELETE SET NULL,
    received_date TIMESTAMPTZ,
    approver_date TIMESTAMPTZ,
    remark        TEXT                NOT NULL DEFAULT '',
    status        incoming_doc_status NOT NULL DEFAULT 'pending',
    dept_id       UUID                REFERENCES departments(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_incoming_docs_doc_id      ON incoming_docs(doc_id);
CREATE INDEX IF NOT EXISTS idx_incoming_docs_sender_id   ON incoming_docs(sender_id);
CREATE INDEX IF NOT EXISTS idx_incoming_docs_receiver_id ON incoming_docs(receiver_id);
CREATE INDEX IF NOT EXISTS idx_incoming_docs_dept_id     ON incoming_docs(dept_id);
CREATE INDEX IF NOT EXISTS idx_incoming_docs_status      ON incoming_docs(status);

COMMENT ON TABLE incoming_docs IS 'Tracks documents received by departments/users, including approval workflow';
