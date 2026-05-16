ALTER TABLE teams
ADD COLUMN IF NOT EXISTS categories JSONB DEFAULT '[]'::jsonb,
ADD COLUMN IF NOT EXISTS city_wide BOOLEAN DEFAULT false;

-- Backfill: copy existing work_area from secretaries to their team's categories
UPDATE teams t
SET categories = COALESCE(
    (
        SELECT jsonb_agg(DISTINCT wa)
        FROM (
            SELECT jsonb_array_elements_text(u.work_area::jsonb) AS wa
            FROM users u
            WHERE u.team_id = t.id
              AND u.type = 'secretary'
              AND u.work_area IS NOT NULL
              AND u.work_area::jsonb != '[]'::jsonb
        ) sub
    ),
    '[]'::jsonb
)
WHERE EXISTS (
    SELECT 1 FROM users u
    WHERE u.team_id = t.id AND u.type = 'secretary' AND u.work_area IS NOT NULL
);