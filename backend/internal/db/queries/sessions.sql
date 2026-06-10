-- name: CreateSession :one
INSERT INTO sessions (user_id, token, expires_at)
VALUES (@user_id, @token, @expires_at)
RETURNING *;

-- name: ExtendSessionByToken :one
UPDATE sessions
SET expires_at = @expires_at
WHERE token = @token
RETURNING *;

-- name: AttachUserToSession :one
-- Used on login: prompote an anon session to a user-bound one
UPDATE sessions
SET user_id = @user_id
WHERE token = @token
RETURNING *;

-- name: DeleteSessionByToken :exec
DELETE FROM sessions
WHERE token = @token;

-- name: DeleteSessionByUserID :exec
DELETE FROM sessions
WHERE user_id = @user_id;

-- name: DeleteOtherSessionsFromUser :exec
DELETE FROM sessions
WHERE user_id = @user_id
  AND id != @session_id;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at < now();
