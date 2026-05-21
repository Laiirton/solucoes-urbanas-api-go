CREATE TABLE IF NOT EXISTS chat_messages (
    id                 BIGSERIAL PRIMARY KEY,
    service_request_id BIGINT NOT NULL REFERENCES service_requests(id) ON DELETE CASCADE,
    sender_id          BIGINT REFERENCES users(id) ON DELETE SET NULL,
    sender_name        TEXT NOT NULL,
    content            TEXT NOT NULL,
    attachments        JSONB,
    created_at         TIMESTAMPTZ DEFAULT NOW(),
    updated_at         TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_service_request_id
    ON chat_messages(service_request_id);

CREATE INDEX IF NOT EXISTS idx_chat_messages_sender_id
    ON chat_messages(sender_id);
