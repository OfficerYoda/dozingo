-- name: GetTwoFactorByUserID :one
SELECT * FROM user_two_factors
WHERE user_id = @user_id;

-- name: CreateTwoFactor :one
INSERT INTO user_two_factors (user_id, totp_secret_encrypted)
VALUES (@user_id, @totp_secret_encrypted)
RETURNING *;

-- name: MarkTwoFactorVerified :one
UPDATE user_two_factors
SET totp_verified_at = now(),
    updated_at = now()
WHERE user_id = @user_id
RETURNING *;

-- name: DeleteTwoFactor :exec
DELETE FROM user_two_factors
WHERE user_id = @user_id;

-- name: UpsertTwoFactor :one
INSERT INTO user_two_factors (user_id, totp_secret_encrypted)
VALUES (@user_id, @totp_secret_encrypted)
ON CONFLICT (user_id) DO UPDATE
    SET totp_secret_encrypted = EXCLUDED.totp_secret_encrypted,
        totp_verified_at      = NULL,
        updated_at            = now()
RETURNING *;

-- name: SetLastUsedCode :exec
UPDATE user_two_factors
SET last_used_code = @code,
    updated_at     = now()
WHERE user_id = @user_id;
