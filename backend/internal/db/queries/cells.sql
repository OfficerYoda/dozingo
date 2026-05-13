-- name: GetCellsByBoardID :many
SELECT * FROM cells
WHERE board_id = $1;

-- name: CreateCell :one
INSERT INTO cells (board_id, content, value)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateCell :one
UPDATE cells
SET content = COALESCE($1, content),
    value = COALESCE($2, value)
WHERE id = $3 AND board_id = $4
RETURNING *;

-- name: DeleteCell :one
DELETE FROM cells
WHERE id = $1 and board_id = $2
RETURNING *;
