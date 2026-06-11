-- name: GetPasswordHashByUserID :one
SELECT password_hash FROM user_passwords
WHERE user_id = @user_id;

-- name: UpsertUserPassword :one
INSERT INTO user_passwords (user_id, password_hash)
VALUES (@user_id, @password_hash)
ON CONFLICT (user_id)
DO UPDATE SET password_hash = EXCLUDED.password_hash
RETURNING *;
