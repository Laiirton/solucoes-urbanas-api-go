CREATE TABLE IF NOT EXISTS news (
    id         BIGSERIAL PRIMARY KEY,
    title      VARCHAR NOT NULL,
    content    TEXT NOT NULL,
    image_urls TEXT[],
    author_id  BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
