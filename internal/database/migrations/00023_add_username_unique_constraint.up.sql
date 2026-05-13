-- Remove duplicate usernames (keep the row with the lowest id)
DELETE FROM users
WHERE id NOT IN (
    SELECT MIN(u.id)
    FROM users u
    GROUP BY u.username
);

-- Add UNIQUE constraint to username
-- This prevents users from sharing the same username
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'uq_users_username'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT uq_users_username UNIQUE (username);
    END IF;
END $$;
