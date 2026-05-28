-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: CreateUser :one
INSERT INTO users (username, email)
VALUES ($1, sqlc.narg('email'))
RETURNING *;

-- name: SetUserEmailVerifiedAt :one
UPDATE users
SET email_verified_at = sqlc.narg('email_verified_at')
WHERE id = $1
RETURNING *;

-- name: DeleteUser :one
DELETE FROM users
WHERE id = $1
RETURNING *;
