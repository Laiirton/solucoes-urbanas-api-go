DROP INDEX IF EXISTS idx_service_requests_team_id;
ALTER TABLE service_requests DROP COLUMN IF EXISTS team_id;
