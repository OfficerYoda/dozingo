-- name: CreateRecoveryCodes :many
INSERT INTO recovery_codes (user_id, code_hash)
VALUES (@user_id, unnest(@code_hashes::text[]))
RETURNING *;

-- name: MarkRecoveryCodeUsed :one
UPDATE recovery_codes
SET used_at    = now(),
    updated_at = now()
WHERE id = @code_id
  AND used_at IS NULL
RETURNING *;

-- name: CountUnusedRecoveryCodesByUserID :one
SELECT COUNT(*) FROM recovery_codes
WHERE user_id = @user_id
  AND used_at IS NULL;

-- name: GetUnusedRecoveryCodesByUserID :many
SELECT * FROM recovery_codes
WHERE user_id = @user_id
  AND used_at IS NULL;

-- name: DeleteRecoveryCodesByUserID :exec
DELETE FROM recovery_codes
WHERE user_id = @user_id;
