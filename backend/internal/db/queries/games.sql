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

-- name: ListGamesByBoard :many
SELECT * FROM games
WHERE board_id = $1
ORDER BY created_at DESC;

-- name: CreateGame :one
INSERT INTO games (player_id, board_id)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateGameStatus :one
UPDATE games
SET status = $1
WHERE id = $2 AND player_id = $3
RETURNING *;

-- name: DeleteGame :one
DELETE FROM games
WHERE id = $1
RETURNING *;
