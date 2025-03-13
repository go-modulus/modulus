-- name: CreateRefreshToken :one
INSERT INTO auth.refresh_token (hash, session_id, identity_id, expires_at)
VALUES (@hash::text, @session_id::uuid, @identity_id, @expires_at)
RETURNING *;

-- name: RevokeRefreshToken :exec
UPDATE auth.refresh_token
SET revoked_at = now()
WHERE hash = $1 AND revoked_at IS NULL;

-- name: RevokeSessionRefreshTokens :exec
UPDATE auth.refresh_token
SET revoked_at = now()
WHERE session_id = $1 AND revoked_at IS NULL;

-- name: RevokeSessionsRefreshTokens :exec
UPDATE auth.refresh_token
SET revoked_at = now()
WHERE session_id = @session_ids::uuid[] AND revoked_at IS NULL;

-- name: GetRefreshTokenByHash :one
SELECT *
FROM auth.refresh_token
WHERE hash = $1;

