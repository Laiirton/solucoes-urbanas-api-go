CREATE TABLE IF NOT EXISTS system_notifications (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT REFERENCES users(id) ON DELETE CASCADE,
    title      VARCHAR(255) NOT NULL,
    body       TEXT NOT NULL,
    type       VARCHAR(50) DEFAULT 'general',
    data       JSONB,
    read_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_system_notifications_user_id ON system_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_system_notifications_type ON system_notifications(type);
CREATE INDEX IF NOT EXISTS idx_system_notifications_read_at ON system_notifications(read_at);
CREATE INDEX IF NOT EXISTS idx_system_notifications_created_at ON system_notifications(created_at DESC);
