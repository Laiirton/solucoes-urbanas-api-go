CREATE TABLE IF NOT EXISTS app_settings (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) UNIQUE NOT NULL,
    value JSONB NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS app_banners (
    id BIGSERIAL PRIMARY KEY,
    image_url TEXT NOT NULL,
    title VARCHAR(255),
    link_url TEXT,
    order_index INT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Initial settings
INSERT INTO app_settings (key, value) VALUES 
('logo_url', '"https://via.placeholder.com/150"'),
('featured_services', '[]'),
('featured_categories', '[]')
ON CONFLICT (key) DO NOTHING;
