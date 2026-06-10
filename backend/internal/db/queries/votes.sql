-- name: GetVotesByBoardID :one
SELECT
    COALESCE(SUM(vote_value), 0)::int AS score,
    COUNT(*)::int                     AS vote_count,
    COALESCE(MAX(CASE WHEN user_id = @user_id THEN vote_value END), 0)::int AS user_vote
FROM votes
WHERE board_id = @board_id;

-- name: UpsertVote :one
INSERT INTO votes (user_id, board_id, vote_value)
VALUES (@user_id, @board_id, @vote_value)
ON CONFLICT (user_id, board_id)
DO UPDATE SET vote_value = EXCLUDED.vote_value
RETURNING *;

-- name: DeleteVote :one
DELETE FROM votes
WHERE user_id = @user_id and board_id = @board_id
RETURNING *;

-- name: ListVotesFromUser :many
SELECT
    v.id          AS vote_id,
    v.vote_value,
    b.id          AS board_id,
    b.title,
    b.description,
    b.size,
    b.author_id   AS board_author_id,
    COALESCE(SUM(all_v.vote_value), 0)::bigint AS score,
    COUNT(all_v.id)::bigint                    AS vote_count,
    COUNT(DISTINCT g.id)::bigint               AS play_count
FROM votes v
JOIN boards b
    ON b.id = v.board_id
LEFT JOIN votes all_v
    ON all_v.board_id = b.id
LEFT JOIN games g
    ON g.board_id = b.id
WHERE v.user_id = @user_id
GROUP BY
    v.id,
    v.vote_value,
    b.id,
    b.title,
    b.description,
    b.size,
    b.author_id
ORDER BY v.created_at DESC;
