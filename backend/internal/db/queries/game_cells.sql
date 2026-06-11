-- name: GetGameCellsByGameID :many
SELECT * FROM game_cells
WHERE game_id = @game_id
ORDER BY position;

-- name: GetGameCellByID :one
SELECT * FROM game_cells
WHERE id = @game_cell_id;

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
SET is_marked = @is_marked
WHERE id = @game_cell_id AND game_id = @game_id
RETURNING *;
