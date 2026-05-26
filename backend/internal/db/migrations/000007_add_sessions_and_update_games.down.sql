-- Drop the session_id column first (depends on the sessions table) and the
-- player_id FK so we can re-tighten its definition below.
ALTER TABLE games DROP CONSTRAINT IF EXISTS games_session_id_fkey;
ALTER TABLE games DROP COLUMN IF EXISTS session_id;
ALTER TABLE games DROP CONSTRAINT IF EXISTS games_player_id_fkey;

-- Restoring NOT NULL on games.player_id is destructive: anonymous games
-- (player_id IS NULL) only became valid because of this migration's up. Drop
-- them so the rollback is runnable without manual cleanup.
DELETE FROM games WHERE player_id IS NULL;
ALTER TABLE games ALTER COLUMN player_id SET NOT NULL;
ALTER TABLE games
    ADD CONSTRAINT games_player_id_fkey
    FOREIGN KEY (player_id) REFERENCES users(id) ON DELETE CASCADE;

DROP TRIGGER IF EXISTS set_updated_at ON sessions;
DROP TABLE IF EXISTS sessions;
