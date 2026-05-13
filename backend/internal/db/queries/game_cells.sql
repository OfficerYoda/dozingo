-- name: GetGameCellsByGameID :many
SELECT * FROM game_cells
WHERE game_id = $1
ORDER BY position;

-- name: CreateGameCells :copyfrom
INSERT INTO game_cells (game_id, cell_id, content, position) VALUES ($1, $2, $3, $4);

-- name: UpdateGameCellMark :one
UPDATE game_cells
SET is_marked = $1
WHERE id = $2 AND game_id = $3
RETURNING *;
