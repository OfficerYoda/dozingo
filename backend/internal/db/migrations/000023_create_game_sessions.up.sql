CREATE TABLE game_sessions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id           UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    started_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at          TIMESTAMPTZ,
    last_heartbeat_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER set_updated_at
BEFORE UPDATE ON game_sessions
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE INDEX idx_game_sessions_open_heartbeat
    ON game_sessions (last_heartbeat_at)
    WHERE ended_at IS NULL;

CREATE INDEX idx_game_sessions_game_id
    ON game_sessions (game_id);

CREATE INDEX idx_game_sessions_effective_end
    ON game_sessions ((COALESCE(ended_at, last_heartbeat_at)));

CREATE INDEX IF NOT EXISTS idx_games_board_id  ON games (board_id);
CREATE INDEX IF NOT EXISTS idx_games_player_id ON games (player_id);
