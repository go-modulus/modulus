-- name: CreateRefreshToken :one
INSERT INTO auth.refresh_token (hash, session_id, data, expires_at)
VALUES (@hash::text, @session_id::uuid, @data::jsonb, @expires_at)
RETURNING *;

-- name: UseRefreshToken :exec
UPDATE auth.refresh_token
SET used_at = @used_at::timestamptz
WHERE hash = $1;

-- name: RevokeRefreshTokens :exec
UPDATE auth.refresh_token
SET revoked_at = @revoked_at::timestamptz
WHERE session_id = $1 AND used_at IS NULL;

-- name: GetRefreshTokenByHash :one
SELECT *
FROM auth.refresh_token
WHERE hash = $1;

