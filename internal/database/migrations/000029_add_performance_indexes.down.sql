-- Migration 000029: Performance indexes + unaccent extension (DOWN)

BEGIN;

DROP INDEX IF EXISTS idx_sr_created_at_status;
DROP INDEX IF EXISTS idx_banners_active;
DROP INDEX IF EXISTS idx_services_category_id;
DROP INDEX IF EXISTS idx_users_cpf;
DROP INDEX IF EXISTS idx_users_type;
DROP INDEX IF EXISTS idx_sn_type_data;
DROP INDEX IF EXISTS idx_sn_user_type;
DROP INDEX IF EXISTS idx_push_tokens_user_id;
DROP INDEX IF EXISTS idx_news_status_created;
DROP INDEX IF EXISTS idx_news_slug;
DROP INDEX IF EXISTS idx_sr_protocol_number;
DROP INDEX IF EXISTS idx_sr_status_created;
DROP INDEX IF EXISTS idx_sr_region_id_category;
DROP INDEX IF EXISTS idx_sr_team_id_category;
DROP INDEX IF EXISTS idx_sr_category;

-- Don't drop the unaccent extension — it may be used elsewhere
-- and is harmless to keep.

COMMIT;
