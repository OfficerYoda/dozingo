-- name: CreateGameSession :one
INSERT INTO game_sessions (game_id)
VALUES (@game_id)
RETURNING *;

-- name: UpdateHeartbeat :one
UPDATE game_sessions
SET last_heartbeat_at = now()
WHERE game_id = @game_id AND ended_at IS NULL
RETURNING *;

-- name: EndGameSessions :execrows
-- Closes any open session for the given game (sets ended_at = now()).
UPDATE game_sessions
SET ended_at = now()
WHERE game_id = @game_id AND ended_at IS NULL;

-- name: CloseStaleSessions :execrows
-- Closes open sessions whose last heartbeat is older than the timeout.
UPDATE game_sessions
SET ended_at = last_heartbeat_at
WHERE ended_at IS NULL
  AND last_heartbeat_at < now() - @timeout::INTERVAL;

-- name: HasOpenGameSession :one
SELECT EXISTS(
    SELECT 1 FROM game_sessions
    WHERE game_id = @game_id AND ended_at IS NULL
) AS has_open;

-- name: GetPlaytimeByGame :one
SELECT COALESCE(
    EXTRACT(EPOCH FROM SUM(COALESCE(ended_at, last_heartbeat_at) - started_at)),
    0
)::bigint AS total_seconds
FROM game_sessions
WHERE game_id = @game_id;

-- name: GetPlaytimeByBoard :one
SELECT COALESCE(
    EXTRACT(EPOCH FROM SUM(COALESCE(gs.ended_at, gs.last_heartbeat_at) - gs.started_at)),
    0
)::bigint AS total_seconds
FROM game_sessions gs
JOIN games g ON g.id = gs.game_id
WHERE g.board_id = @board_id;

-- name: GetPlaytimeByPlayer :one
SELECT COALESCE(
    EXTRACT(EPOCH FROM SUM(COALESCE(gs.ended_at, gs.last_heartbeat_at) - gs.started_at)),
    0
)::bigint AS total_seconds
FROM game_sessions gs
JOIN games g ON g.id = gs.game_id
WHERE g.player_id = @player_id;

-- name: GetTotalPlaytime :one
SELECT COALESCE(
    EXTRACT(EPOCH FROM SUM(
        LEAST(COALESCE(ended_at, last_heartbeat_at), now())
      - GREATEST(started_at, now() - @period::INTERVAL)
    )),
    0
)::bigint AS total_seconds
FROM game_sessions
WHERE COALESCE(ended_at, last_heartbeat_at) >= now() - @period::INTERVAL;
