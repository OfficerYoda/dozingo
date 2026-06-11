-- name: GetUserByID :one
SELECT * FROM users
WHERE id = @user_id;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = @username;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = @email;

-- name: CreateUser :one
INSERT INTO users (username, email)
VALUES (@username, sqlc.narg('email'))
RETURNING *;

-- name: SetUserEmailVerifiedAt :one
UPDATE users
SET email_verified_at = sqlc.narg('email_verified_at')
WHERE id = @user_id
RETURNING *;

-- name: UpdateUser :one
-- Tri-state PATCH:
--   * username: NULL means "leave alone", non-NULL means "set"
--   * email: NULL means "leave alone", non-NULL means "set"
--   * clear_email: true means "clear email and reset email_verified_at"
UPDATE users
SET
    username = COALESCE(sqlc.narg('username'), username),
    email = CASE
        WHEN sqlc.arg('clear_email')::bool THEN NULL
        WHEN sqlc.narg('email')::text IS NOT NULL THEN sqlc.narg('email')
        ELSE email
    END,
    email_verified_at = CASE
        WHEN sqlc.arg('clear_email')::bool THEN NULL
        WHEN sqlc.narg('email')::text IS NOT NULL THEN NULL
        ELSE email_verified_at
    END
WHERE id = @user_id
RETURNING *;

-- name: DeleteUser :one
DELETE FROM users
WHERE id = @user_id
RETURNING *;

-- name: SetAvatar :one
UPDATE users
SET avatar_key = @avatar_key
WHERE id = @user_id
RETURNING *;

-- name: ListInUseAvatarKeys :many
SELECT DISTINCT avatar_key FROM users
WHERE avatar_key <> '';
