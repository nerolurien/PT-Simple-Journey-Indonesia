ALTER TABLE tickets ADD COLUMN IF NOT EXISTS first_responded_at TIMESTAMPTZ NULL;

CREATE TABLE IF NOT EXISTS ticket_attachments (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    uploader_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    original_name VARCHAR(255) NOT NULL,
    stored_name VARCHAR(255) NOT NULL UNIQUE,
    mime_type VARCHAR(128) NOT NULL,
    size_bytes BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_attachments_ticket_id ON ticket_attachments(ticket_id);
