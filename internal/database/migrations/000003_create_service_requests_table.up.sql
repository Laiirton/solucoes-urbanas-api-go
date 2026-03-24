CREATE TYPE IF NOT EXISTS service_request_status AS ENUM ('pending', 'in_progress', 'completed', 'cancelled');

CREATE TABLE IF NOT EXISTS service_requests (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT REFERENCES users(id) ON DELETE SET NULL,
    service_id      BIGINT NOT NULL REFERENCES services(id) ON DELETE RESTRICT,
    protocol_number TEXT UNIQUE,
    service_title   TEXT NOT NULL,
    category        TEXT NOT NULL,
    request_data    JSONB NOT NULL DEFAULT '{}',
    attachments     JSONB,
    status          service_request_status NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
