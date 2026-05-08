DROP INDEX IF EXISTS idx_service_requests_region_id;
DROP INDEX IF EXISTS idx_teams_region_id;

ALTER TABLE service_requests DROP COLUMN IF EXISTS region_id;

ALTER TABLE teams ADD COLUMN IF NOT EXISTS service_category VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE teams DROP COLUMN IF EXISTS region_id;

DROP TABLE IF EXISTS regions;
