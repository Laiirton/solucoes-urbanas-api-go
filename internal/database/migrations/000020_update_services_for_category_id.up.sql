-- Add category_id column if not exists (PostgreSQL < 9.3 compatible)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'services' AND column_name = 'category_id'
    ) THEN
        ALTER TABLE services ADD COLUMN category_id BIGINT;
    END IF;
END $$;

-- Update existing records to set category_id based on category name
UPDATE services 
SET category_id = c.id 
FROM categories c 
WHERE services.category = c.name
AND services.category_id IS NULL;

-- Drop existing constraint if any (from migration 19 which created it without ON DELETE SET NULL)
ALTER TABLE services DROP CONSTRAINT IF EXISTS fk_services_category;

-- Re-create with ON DELETE SET NULL
ALTER TABLE services 
ADD CONSTRAINT fk_services_category 
FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_services_category_id ON services(category_id);