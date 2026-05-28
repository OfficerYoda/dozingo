-- name: CreateVerificationToken :one
INSERT INTO verification_tokens (user_id, token, type, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetVerificationTokenByToken :one
SELECT * FROM verification_tokens
WHERE token = $1 
  AND expires_at > now();

-- name: GetValidTokenForUser :one
-- Fetches an unexpired token of a specific type for a user
SELECT * FROM verification_tokens
WHERE user_id = $1 
  AND type = $2 
  AND expires_at > now();

-- name: DeleteVerificationToken :exec
DELETE FROM verification_tokens
WHERE token = $1;

-- name: DeleteExpiredTokens :exec
DELETE FROM verification_tokens
WHERE expires_at < now();
