ALTER TABLE games DROP CONSTRAINT IF EXISTS games_session_id_fkey;
ALTER TABLE games DROP COLUMN IF EXISTS session_id;
ALTER TABLE games DROP CONSTRAINT IF EXISTS games_player_id_fkey;
ALTER TABLE games ALTER COLUMN player_id SET NOT NULL;
ALTER TABLE games
  ADD CONSTRAINT games_player_id_fkey
  FOREIGN KEY (player_id) REFERENCES users(id) ON DELETE CASCADE;

DROP TRIGGER IF EXISTS set_updated_at ON sessions;
DROP TABLE IF EXISTS sessions;
