-- name: GetUserForPasswordLogin :one
SELECT u.id, u.username, u.email, u.avatar_key, up.password_hash
FROM users u
INNER JOIN user_passwords up ON up.user_id = u.id
WHERE u.username = @username;

-- name: GetSessionUserByToken :one
-- user_id may be NULL for anon sessions
SELECT 
  s.id AS session_id,
  s.user_id,
  s.token,
  s.expires_at,
  s.two_fa_pending,
  u.username,
  u.email,
  u.email_verified_at,
  u.avatar_key
FROM sessions s
LEFT JOIN users u ON u.id = s.user_id
WHERE s.token = @token
  AND s.expires_at > now();
