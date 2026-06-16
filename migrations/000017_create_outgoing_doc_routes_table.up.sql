-- Migration: create outgoing_doc_routes table
-- Stores the ordered department route an outgoing document must flow through.
-- sequence_order 1 = owner department, then selected departments in order,
-- then Director Office (if requested). incoming_doc_id is set lazily once that
-- step's incoming document is created (only after the previous step is approved).

CREATE TABLE IF NOT EXISTS outgoing_doc_routes (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    outgoing_doc_id UUID        NOT NULL REFERENCES outgoing_docs(id) ON DELETE CASCADE,
    dept_id         UUID        NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    sequence_order  INT         NOT NULL,
    incoming_doc_id UUID        REFERENCES incoming_docs(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (outgoing_doc_id, sequence_order)
);

CREATE INDEX IF NOT EXISTS idx_outgoing_doc_routes_outgoing_doc_id ON outgoing_doc_routes(outgoing_doc_id);
CREATE INDEX IF NOT EXISTS idx_outgoing_doc_routes_dept_id         ON outgoing_doc_routes(dept_id);
CREATE INDEX IF NOT EXISTS idx_outgoing_doc_routes_incoming_doc_id ON outgoing_doc_routes(incoming_doc_id);

COMMENT ON TABLE outgoing_doc_routes IS 'Ordered department route for an outgoing document; drives sequential approval flow';
