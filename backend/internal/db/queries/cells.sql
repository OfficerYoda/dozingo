-- name: GetCellsByBoardID :many
SELECT * FROM cells
WHERE board_id = @board_id;

-- name: GetCellByID :one
SELECT * FROM cells
WHERE id = @cell_id;

-- name: ListCellsByIDs :many
SELECT * FROM cells
WHERE id = ANY(@cell_ids::uuid[])
  AND board_id = @board_id;

-- name: CreateCell :one
INSERT INTO cells (board_id, content, value)
VALUES (@board_id, @content, @value)
RETURNING *;

-- name: UpdateCell :one
UPDATE cells
SET content = COALESCE(sqlc.narg('content'), content),
    value = COALESCE(sqlc.narg('value'), value)
WHERE id = sqlc.arg('cell_id') AND board_id = sqlc.arg('board_id')
RETURNING *;

-- name: DeleteCell :one
DELETE FROM cells
WHERE id = @cell_id and board_id = @board_id
RETURNING *;
