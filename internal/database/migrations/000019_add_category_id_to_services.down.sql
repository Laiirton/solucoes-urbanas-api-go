ALTER TABLE services DROP CONSTRAINT IF EXISTS fk_services_category;
ALTER TABLE services DROP COLUMN IF EXISTS category_id;
DROP INDEX IF EXISTS idx_services_category_id;