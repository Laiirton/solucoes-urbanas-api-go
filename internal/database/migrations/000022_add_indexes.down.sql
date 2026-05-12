-- Migration 000022: Down — Drop all indexes added in the up migration

BEGIN;

DROP INDEX IF EXISTS idx_sr_user_id;
DROP INDEX IF EXISTS idx_sr_service_id;
DROP INDEX IF EXISTS idx_users_team_id;
DROP INDEX IF EXISTS idx_ratings_service_id;
DROP INDEX IF EXISTS idx_sa_request_id;
DROP INDEX IF EXISTS idx_sr_created_at;
DROP INDEX IF EXISTS idx_sr_status;
DROP INDEX IF EXISTS idx_services_category;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_teams_region_id;
DROP INDEX IF EXISTS idx_regions_neighborhoods_gin;
DROP INDEX IF EXISTS idx_sr_user_id_id;
DROP INDEX IF EXISTS idx_sr_team_id_id;
DROP INDEX IF EXISTS idx_sr_region_id_id;
DROP INDEX IF EXISTS idx_sr_service_id_created;
DROP INDEX IF EXISTS idx_sr_status_id;

-- Re-create the dropped index (even though unused, restore for backwards compatibility)
CREATE INDEX IF NOT EXISTS idx_users_profile_image ON users(profile_image_url);

COMMIT;
