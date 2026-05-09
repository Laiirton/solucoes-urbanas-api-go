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

-- Add foreign key constraint if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE table_name = 'services' AND constraint_name = 'fk_services_category'
    ) THEN
        ALTER TABLE services 
        ADD CONSTRAINT fk_services_category 
        FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL;
    END IF;
END $$;

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_services_category_id ON services(category_id);