ALTER TABLE service_requests
ADD COLUMN IF NOT EXISTS team_id BIGINT REFERENCES teams(id) ON DELETE SET NULL;

-- Backfill: match existing requests to the team responsible for their category
UPDATE service_requests sr
SET team_id = t.id
FROM teams t
WHERE sr.team_id IS NULL
  AND sr.category = t.service_category;

CREATE INDEX IF NOT EXISTS idx_service_requests_team_id ON service_requests(team_id);
