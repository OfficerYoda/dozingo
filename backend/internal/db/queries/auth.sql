-- name: GetUserForPasswordLogin :one
SELECT u.id, u.username, u.email, up.password_hash
FROM users u
INNER JOIN user_passwords up ON up.user_id = u.id
WHERE u.username = $1;
