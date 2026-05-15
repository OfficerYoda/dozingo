CREATE TABLE sessions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON sessions
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX ON sessions(expires_at);

ALTER TABLE games DROP CONSTRAINT games_player_id_fkey;
ALTER TABLE games ALTER COLUMN player_id DROP NOT NULL;
ALTER TABLE games
  ADD CONSTRAINT games_player_id_fkey
  FOREIGN KEY (player_id) REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE games
  ADD COLUMN session_id UUID REFERENCES sessions(id) ON DELETE SET NULL;
