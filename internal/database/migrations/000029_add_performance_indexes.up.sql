-- Migration 000029: Performance indexes + unaccent extension
-- Adds missing indexes for query performance and installs unaccent extension
-- to replace the monster TRANSLATE expressions used for accent-insensitive matching.

BEGIN;

-- ============================================================
-- Install unaccent extension (safe idempotent)
-- This replaces TRANSLATE('áàâãä...', 'aaaaa...') everywhere
-- ============================================================
CREATE EXTENSION IF NOT EXISTS unaccent;

-- ============================================================
-- Indexes on service_requests
-- ============================================================

-- Category filter (used in 8+ query patterns: home stats, team filter, map locations, dashboard)
CREATE INDEX IF NOT EXISTS idx_sr_category ON service_requests(category);

-- Composite: team_id + category (team-scoped queries with category fallback)
CREATE INDEX IF NOT EXISTS idx_sr_team_id_category ON service_requests(team_id, category);

-- Composite: region_id + category (region-scoped queries)
CREATE INDEX IF NOT EXISTS idx_sr_region_id_category ON service_requests(region_id, category);

-- Status + created_at (alert queries: "pending/in_progress older than 3 days")
CREATE INDEX IF NOT EXISTS idx_sr_status_created ON service_requests(status, created_at);

-- Protocol number generation (avoids sequential scan on protocol_number IS NULL)
CREATE INDEX IF NOT EXISTS idx_sr_protocol_number ON service_requests(protocol_number) WHERE protocol_number IS NOT NULL;

-- ============================================================
-- Indexes on news
-- ============================================================

-- Slug lookup (GetNewsBySlug)
CREATE INDEX IF NOT EXISTS idx_news_slug ON news(slug);

-- Status filter + sort (ListNews with status and date filters)
CREATE INDEX IF NOT EXISTS idx_news_status_created ON news(status, created_at DESC);

-- ============================================================
-- Indexes on push tokens & notifications
-- ============================================================

-- User push token lookup (SendToUser, DeletePushToken)
CREATE INDEX IF NOT EXISTS idx_push_tokens_user_id ON user_push_tokens(user_id);

-- System notification filtering (List by user_id + type)
CREATE INDEX IF NOT EXISTS idx_sn_user_type ON system_notifications(user_id, type);

-- System notification cleanup by type + ref_id (DeleteByTypeAndRefID)
CREATE INDEX IF NOT EXISTS idx_sn_type_data ON system_notifications(type, (data->>'news_id'));

-- ============================================================
-- Indexes on users & services
-- ============================================================

-- User role filter (ListUsers by type for admin dashboards)
CREATE INDEX IF NOT EXISTS idx_users_type ON users(type);

-- User CPF lookup (duplicate check in CreateUser)
CREATE INDEX IF NOT EXISTS idx_users_cpf ON users(cpf);

-- Service by category_id (ListServicesByCategoryID)
CREATE INDEX IF NOT EXISTS idx_services_category_id ON services(category_id);

-- ============================================================
-- Indexes on app configs
-- ============================================================

-- Active banners (GetBanners WHERE is_active = TRUE)
CREATE INDEX IF NOT EXISTS idx_banners_active ON app_banners(is_active) WHERE is_active = TRUE;

-- ============================================================
-- Composite index for common date-range + status queries
-- ============================================================

-- Home dashboard: created_at + status (used in computeVolume7d, computeRecentRequests)
CREATE INDEX IF NOT EXISTS idx_sr_created_at_status ON service_requests(created_at DESC, status);

COMMIT;
