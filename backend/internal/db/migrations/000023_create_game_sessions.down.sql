DROP INDEX IF EXISTS idx_games_player_id;
DROP INDEX IF EXISTS idx_games_board_id;
DROP TRIGGER IF EXISTS set_updated_at ON game_sessions;
DROP TABLE IF EXISTS game_sessions;
