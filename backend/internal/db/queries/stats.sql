-- name: GetRecentStats :one
SELECT
    COALESCE(SUM(bingo_count) FILTER (WHERE updated_at >= now() - @period::INTERVAL), 0) AS bingos,
    COUNT(*) FILTER (WHERE created_at >= now() - @period::INTERVAL)                       AS games,
    (SELECT COUNT(*) FROM boards WHERE created_at >= now() - @period::INTERVAL)            AS boards,
    (SELECT COUNT(*) FROM cells  WHERE created_at >= now() - @period::INTERVAL)            AS cells
FROM games;
