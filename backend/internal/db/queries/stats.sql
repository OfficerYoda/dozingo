-- name: GetRecentStats :one
SELECT
    COUNT(*) FILTER (WHERE status = 'completed' AND updated_at >= now() - @period::INTERVAL) AS bingos,
    COUNT(*) FILTER (WHERE created_at >= now() - @period::INTERVAL)                           AS games,
    (SELECT COUNT(*) FROM boards WHERE created_at >= now() - @period::INTERVAL)               AS boards,
    (SELECT COUNT(*) FROM cells  WHERE created_at >= now() - @period::INTERVAL)               AS cells
FROM games;
