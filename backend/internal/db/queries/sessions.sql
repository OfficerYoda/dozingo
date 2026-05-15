-- name: CreateSession :one
INSERT INTO sessions (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSessionByToken :one
-- user_id may be NULL for anon sessions
SELECT 
  s.id AS session_id,
  s.user_id,
  s.token,
  s.expires_at,
  u.id AS user_id,
  u.username,
  u.email
FROM sessions s
LEFT JOIN users u ON u.id = s.user_id
WHERE s.token = $1
  AND s.expires_at > now();

-- name: ExtendSessionByToken :one
UPDATE sessions
SET expires_at = $2
WHERE token = $1
RETURNING *;

-- name: AttachUserToSession :one
-- Used on register: prompote an anon session to a user-bound one
UPDATE sessions
SET user_id = $2
WHERE token = $1
RETURNING *;

-- name: DeleteSessionByToken :exec
DELETE FROM sessions
WHERE token = $1;

-- name: DeleteSessionsByUserID :exec
DELETE FROM sessions
WHERE user_id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at < now();
