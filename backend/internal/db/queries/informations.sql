-- name: GetSecurityInformation :one
SELECT
    up.updated_at AS password_last_changed_at,
    (
        SELECT s.created_at
        FROM sessions s
        WHERE s.user_id = @user_id
          AND s.expires_at > now()
        ORDER BY s.created_at DESC
        LIMIT 1
    ) AS last_login_at,
    (
        SELECT COUNT(*)
        FROM sessions s
        WHERE s.user_id = @user_id
          AND s.expires_at > now()
    ) AS active_sessions_count
FROM user_passwords up
WHERE up.user_id = @user_id;
