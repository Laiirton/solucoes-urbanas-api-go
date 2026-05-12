-- Migration 000022: Add missing indexes for query performance
-- This migration adds indexes on foreign key columns and frequently filtered columns
-- that were missing from the original schema.

BEGIN;

-- ============================================================
-- HIGH PRIORITY — FK columns without indexes (full scan in JOINs)
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_sr_user_id ON service_requests(user_id);

CREATE INDEX IF NOT EXISTS idx_sr_service_id ON service_requests(service_id);

CREATE INDEX IF NOT EXISTS idx_users_team_id ON users(team_id);

CREATE INDEX IF NOT EXISTS idx_ratings_service_id ON service_ratings(service_id);

CREATE INDEX IF NOT EXISTS idx_sa_request_id ON service_attendances(service_request_id);

-- ============================================================
-- HIGH PRIORITY — Frequently filtered columns
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_sr_created_at ON service_requests(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_sr_status ON service_requests(status);

CREATE INDEX IF NOT EXISTS idx_services_category ON services(category);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

CREATE INDEX IF NOT EXISTS idx_teams_region_id ON teams(region_id);

-- ============================================================
-- GIN index for JSONB containment search on regions.neighborhoods
-- ============================================================

CREATE INDEX IF NOT EXISTS idx_regions_neighborhoods_gin ON regions USING GIN (neighborhoods);

-- ============================================================
-- Composite indexes for common query patterns
-- ============================================================

-- WHERE user_id = $1 ORDER BY id DESC
CREATE INDEX IF NOT EXISTS idx_sr_user_id_id ON service_requests(user_id, id DESC);

-- WHERE team_id = $1 ORDER BY id DESC
CREATE INDEX IF NOT EXISTS idx_sr_team_id_id ON service_requests(team_id, id DESC);

-- WHERE region_id = $1 ORDER BY id DESC
CREATE INDEX IF NOT EXISTS idx_sr_region_id_id ON service_requests(region_id, id DESC);

-- WHERE service_id = $1 ORDER BY created_at DESC
CREATE INDEX IF NOT EXISTS idx_sr_service_id_created ON service_requests(service_id, created_at DESC);

-- WHERE status = $1 ORDER BY id DESC
CREATE INDEX IF NOT EXISTS idx_sr_status_id ON service_requests(status, id DESC);

-- ============================================================
-- Drop unused index
-- ============================================================

DROP INDEX IF EXISTS idx_users_profile_image;

COMMIT;
