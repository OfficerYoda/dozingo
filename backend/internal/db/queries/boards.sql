-- name: GetBoards :many
SELECT * FROM boards
ORDER BY created_at DESC;

-- name: GetBoardByID :one
SELECT * FROM boards
WHERE id = $1;

-- name: CreateBoard :one
INSERT INTO boards (title, size, author_id, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: DeleteBoard :one
DELETE FROM boards
WHERE id = $1
RETURNING *;

-- name: GetTotalGamesPlayedForBoard :one
SELECT
    b.id AS board_id,
    b.title AS board_title,
    COUNT(g.id) AS total_games
FROM boards b
LEFT JOIN games g ON g.board_id = b.id
WHERE b.id = $1
GROUP BY b.id, b.title;
