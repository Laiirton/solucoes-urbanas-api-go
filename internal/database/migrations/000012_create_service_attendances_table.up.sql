CREATE TABLE IF NOT EXISTS service_attendances (
    id                 BIGSERIAL PRIMARY KEY,
    service_request_id BIGINT NOT NULL REFERENCES service_requests(id) ON DELETE CASCADE,
    attended_by        BIGINT REFERENCES users(id) ON DELETE SET NULL,
    notes              TEXT,
    attachments        JSONB,
    created_at         TIMESTAMPTZ DEFAULT NOW(),
    updated_at         TIMESTAMPTZ DEFAULT NOW()
);

-- Add a trigger to update the updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_service_attendances_updated_at
    BEFORE UPDATE ON service_attendances
    FOR EACH ROW
    EXECUTE PROCEDURE update_updated_at_column();
