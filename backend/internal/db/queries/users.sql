-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (username, email)
VALUES ($1, sqlc.narg('email'))
RETURNING *;

-- name: SetUserEmailVerifiedAt :one
UPDATE users
SET email_verified_at = sqlc.narg('email_verified_at')
WHERE id = $1
RETURNING *;

-- name: UpdateUser :one
-- Tri-state PATCH:
--   * username: NULL means "leave alone", non-NULL means "set"
--   * email_set=false means "leave email/email_verified_at alone";
--     email_set=true writes whatever's in email (NULL = clear) and
--     resets email_verified_at to NULL.
UPDATE users
SET
    username = COALESCE(sqlc.narg('username'), username),
    email = CASE
        WHEN sqlc.arg('email_set')::bool THEN sqlc.narg('email')
        ELSE email
    END,
    email_verified_at = CASE
        WHEN sqlc.arg('email_set')::bool THEN NULL
        ELSE email_verified_at
    END
WHERE id = $1
RETURNING *;

-- name: DeleteUser :one
DELETE FROM users
WHERE id = $1
RETURNING *;

-- name: SetAvatar :one
UPDATE users
SET avatar_key = @avatar_key
WHERE id = @user_id
RETURNING *;
