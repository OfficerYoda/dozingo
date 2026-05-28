-- Add email_verified_at to users
ALTER TABLE users ADD COLUMN email_verified_at TIMESTAMPTZ;

-- Convert games.status from TEXT to game_status enum
CREATE TYPE game_status AS ENUM ('active', 'completed', 'abandoned');

ALTER TABLE games ALTER COLUMN status DROP DEFAULT;
ALTER TABLE games DROP CONSTRAINT IF EXISTS chk_game_status;
ALTER TABLE games DROP CONSTRAINT IF EXISTS games_status_check;
ALTER TABLE games
    ALTER COLUMN status TYPE game_status
    USING status::game_status;
ALTER TABLE games ALTER COLUMN status SET DEFAULT 'active';
