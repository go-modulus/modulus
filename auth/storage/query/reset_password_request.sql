-- name: GetReadyForUseResetPasswordRequest :one
SELECT *
FROM auth.reset_password_request
WHERE status = 'active' AND account_id = $1;

-- name: ExpireResetPasswordRequest :exec
UPDATE auth.reset_password_request
SET status  = 'expired'
WHERE id = $1;

-- name: CreateResetPasswordRequest :one
INSERT INTO auth.reset_password_request
    (id, account_id, token)
VALUES (@id, @account_id, @token)
RETURNING *;

-- name: UpdateLastSentRequest :exec
UPDATE auth.reset_password_request
SET last_send_at = now()
WHERE id = $1;

-- name: GetResetPasswordRequestByToken :one
SELECT *
FROM auth.reset_password_request
WHERE token = $1;

-- name: UseResetPasswordRequest :exec
UPDATE auth.reset_password_request
SET status  = 'used',
    used_at = now()
WHERE id = $1;

-- name: DeleteResetPasswordRequests :exec
DELETE FROM auth.reset_password_request
WHERE account_id = $1;
