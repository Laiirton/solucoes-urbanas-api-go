-- Add lat/lon/geocoded_address (safe: IF NOT EXISTS for manual additions)
ALTER TABLE service_requests
    ADD COLUMN IF NOT EXISTS latitude        DECIMAL,
    ADD COLUMN IF NOT EXISTS longitude       DECIMAL,
    ADD COLUMN IF NOT EXISTS geocoded_address TEXT;

-- Enforce 1:1 team-to-region mapping
ALTER TABLE teams ADD CONSTRAINT uq_teams_region_id UNIQUE (region_id);
