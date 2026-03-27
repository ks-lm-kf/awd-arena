-- Add SSH user and password fields to team_containers table
ALTER TABLE team_containers ADD COLUMN IF NOT EXISTS ssh_user VARCHAR(50) DEFAULT 'awd';
ALTER TABLE team_containers ADD COLUMN IF NOT EXISTS ssh_password VARCHAR(100);

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_team_containers_game_team ON team_containers(game_id, team_id);
