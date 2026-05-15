-- name: GetGames :many
SELECT * FROM games
ORDER BY created_at DESC;

-- name: GetGameByID :one
SELECT * FROM games
WHERE id = $1;

-- name: ListGamesByPlayer :many
SELECT * FROM games
WHERE player_id = $1
ORDER BY created_at DESC;

-- name: ListGamesBySession :many
SELECT * FROM games
WHERE session_id = $1
ORDER BY created_at DESC;

-- name: ListGamesByBoard :many
SELECT * FROM games
WHERE board_id = $1
ORDER BY created_at DESC;

-- name: CreateGame :one
INSERT INTO games (player_id, session_id, board_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateGameStatus :one
-- Authorize by either player_id (logged-in) or session_id (anon)
UPDATE games
SET status = $1
WHERE id = $2
  AND (
        (sqlc.narg('player_id')::uuid IS NOT NULL AND player_id = sqlc.narg('player_id'))
     OR (sqlc.narg('player_id')::uuid IS NULL     AND session_id = sqlc.narg('session_id'))
  )
RETURNING *;

-- name: DeleteGame :one
DELETE FROM games
WHERE id = $1
RETURNING *;
