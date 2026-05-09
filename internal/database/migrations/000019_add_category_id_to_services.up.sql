ALTER TABLE services ADD COLUMN IF NOT EXISTS category_id BIGINT;

UPDATE services 
SET category_id = c.id 
FROM categories c 
WHERE services.category = c.name;

ALTER TABLE services 
ADD CONSTRAINT IF NOT EXISTS fk_services_category 
FOREIGN KEY (category_id) REFERENCES categories(id);

-- Index for performance
CREATE INDEX IF NOT EXISTS idx_services_category_id ON services(category_id);