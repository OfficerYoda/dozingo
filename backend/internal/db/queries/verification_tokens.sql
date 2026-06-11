-- name: CreateVerificationToken :one
INSERT INTO verification_tokens (user_id, token, type, expires_at)
VALUES (@user_id, @token, @type, @expires_at)
RETURNING *;

-- name: GetVerificationTokenByToken :one
SELECT * FROM verification_tokens
WHERE token = @token 
  AND expires_at > now();

-- name: GetValidTokenForUser :one
-- Fetches an unexpired token of a specific type for a user
SELECT * FROM verification_tokens
WHERE user_id = @user_id 
  AND type = @type 
  AND expires_at > now();

-- name: DeleteVerificationToken :exec
DELETE FROM verification_tokens
WHERE token = @token;

-- name: DeleteExpiredTokens :exec
DELETE FROM verification_tokens
WHERE expires_at < now();
