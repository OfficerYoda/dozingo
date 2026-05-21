-- name: GetGameCellsByGameID :many
SELECT * FROM game_cells
WHERE game_id = $1
ORDER BY position;

-- name: CreateGameCells :many
INSERT INTO game_cells (game_id, cell_id, content, position)
SELECT
    unnest(@game_ids::uuid[]),
    unnest(@cell_ids::uuid[]),
    unnest(@contents::text[]),
    unnest(@positions::int[])
RETURNING *;

-- name: UpdateGameCellMark :one
UPDATE game_cells
SET is_marked = $1
WHERE id = $2 AND game_id = $3
RETURNING *;
