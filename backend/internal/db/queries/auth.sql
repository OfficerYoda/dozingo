-- name: GetUserForPasswordLogin :one
SELECT u.id, u.username, u.email, up.password_hash
FROM users u
INNER JOIN user_passwords up ON up.user_id = u.id
WHERE u.username = $1;

-- name: GetSessionUserByToken :one
-- user_id may be NULL for anon sessions
SELECT 
  s.id AS session_id,
  s.user_id,
  s.token,
  s.expires_at,
  u.username,
  u.email
FROM sessions s
LEFT JOIN users u ON u.id = s.user_id
WHERE s.token = $1
  AND s.expires_at > now();
