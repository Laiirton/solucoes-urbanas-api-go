ALTER TABLE service_requests
    DROP COLUMN IF EXISTS latitude,
    DROP COLUMN IF EXISTS longitude,
    DROP COLUMN IF EXISTS geocoded_address;

ALTER TABLE teams DROP CONSTRAINT IF EXISTS uq_teams_region_id;
