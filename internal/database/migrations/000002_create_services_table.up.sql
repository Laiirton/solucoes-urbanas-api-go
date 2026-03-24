CREATE TABLE IF NOT EXISTS services (
    id          BIGSERIAL PRIMARY KEY,
    title       VARCHAR   NOT NULL,
    description TEXT,
    category    VARCHAR   NOT NULL,
    is_active   BOOLEAN   NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);
