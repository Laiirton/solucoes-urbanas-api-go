-- Add category_id column to services table
ALTER TABLE services ADD COLUMN IF NOT EXISTS category_id BIGINT;

-- Update existing records to set category_id based on category name
UPDATE services 
SET category_id = c.id 
FROM categories c 
WHERE services.category = c.name
AND services.category_id IS NULL;

-- Add foreign key constraint
ALTER TABLE services 
ADD CONSTRAINT IF NOT EXISTS fk_services_category 
FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_services_category_id ON services(category_id);