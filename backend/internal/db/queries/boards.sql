-- name: GetBoards :many
SELECT * FROM boards
ORDER BY created_at DESC;

-- name: GetBoardByID :one
SELECT * FROM boards
WHERE id = @board_id;

-- name: CreateBoard :one
INSERT INTO boards (title, size, author_id, description)
VALUES (@title, @size, @author_id, @description)
RETURNING *;

-- name: DeleteBoard :one
DELETE FROM boards
WHERE id = @board_id
RETURNING *;

-- name: GetTotalGamesPlayedForBoard :one
SELECT
    b.id AS board_id,
    b.title AS board_title,
    COUNT(g.id) AS total_games
FROM boards b
LEFT JOIN games g ON g.board_id = b.id
WHERE b.id = @board_id
GROUP BY b.id, b.title;
