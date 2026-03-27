-- Add password change tracking fields
ALTER TABLE users
ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT true;

-- Update existing users to not require password change
-- (You can remove this if you want all existing users to change password on next login)
UPDATE users SET must_change_password = false, password_changed_at = NOW() WHERE must_change_password = true;

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_users_must_change_password ON users(must_change_password);
