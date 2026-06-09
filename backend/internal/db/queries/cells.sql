-- name: GetCellsByBoardID :many
SELECT * FROM cells
WHERE board_id = $1;

-- name: GetCellByID :one
SELECT * FROM cells
WHERE id = $1;

-- name: CreateCell :one
INSERT INTO cells (board_id, content, value)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateCell :one
UPDATE cells
SET content = COALESCE(sqlc.narg('content'), content),
    value = COALESCE(sqlc.narg('value'), value)
WHERE id = sqlc.arg('cell_id') AND board_id = sqlc.arg('board_id')
RETURNING *;

-- name: DeleteCell :one
DELETE FROM cells
WHERE id = $1 and board_id = $2
RETURNING *;
