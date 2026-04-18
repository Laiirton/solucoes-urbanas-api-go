-- Add new columns for a professional news system
ALTER TABLE news ADD COLUMN IF NOT EXISTS slug VARCHAR UNIQUE;
ALTER TABLE news ADD COLUMN IF NOT EXISTS summary TEXT;
ALTER TABLE news ADD COLUMN IF NOT EXISTS status VARCHAR DEFAULT 'draft';
ALTER TABLE news ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ;
ALTER TABLE news ADD COLUMN IF NOT EXISTS category VARCHAR;
ALTER TABLE news ADD COLUMN IF NOT EXISTS tags TEXT[];

-- Convert content to JSONB if it's not already
-- Using to_jsonb(content) safely converts existing plain text into a valid JSON string
ALTER TABLE news ALTER COLUMN content TYPE JSONB USING to_jsonb(content);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_news_slug ON news(slug);
CREATE INDEX IF NOT EXISTS idx_news_status ON news(status);
CREATE INDEX IF NOT EXISTS idx_news_published_at ON news(published_at);
