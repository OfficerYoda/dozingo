-- name: CreateSession :one
INSERT INTO sessions (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ExtendSessionByToken :one
UPDATE sessions
SET expires_at = $2
WHERE token = $1
RETURNING *;

-- name: AttachUserToSession :one
-- Used on login: prompote an anon session to a user-bound one
UPDATE sessions
SET user_id = $2
WHERE token = $1
RETURNING *;

-- name: DeleteSessionByToken :exec
DELETE FROM sessions
WHERE token = $1;

-- name: DeleteSessionByUserID :exec
DELETE FROM sessions
WHERE user_id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at < now();
