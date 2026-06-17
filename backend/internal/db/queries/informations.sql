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
    ) AS active_sessions_count,
    EXISTS (
        SELECT 1 FROM user_two_factors utf
        WHERE utf.user_id = @user_id
          AND utf.totp_verified_at IS NOT NULL
    ) AS two_factor_enabled,
    (
        SELECT COUNT(*)
        FROM recovery_codes rc
        WHERE rc.user_id = @user_id
          AND rc.used_at IS NULL
    ) AS unused_recovery_keys
FROM user_passwords up
WHERE up.user_id = @user_id;
