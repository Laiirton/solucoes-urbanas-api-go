-- Remove services table and the relationship in service_requests
-- Using IF EXISTS to avoid errors if the column/table is already gone or never existed

DROP TABLE IF EXISTS services CASCADE;

ALTER TABLE service_requests DROP COLUMN IF EXISTS service_id;
