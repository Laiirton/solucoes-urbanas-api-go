-- Add profile_image_url column to users table
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS profile_image_url VARCHAR NULL;

-- Add index for faster lookups if needed
CREATE INDEX IF NOT EXISTS idx_users_profile_image ON users(profile_image_url);
