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
WHERE session_id = ANY(@session_ids::uuid[]) AND revoked_at IS NULL;

-- name: FindRefreshTokenByHash :one
SELECT *
FROM auth.refresh_token
WHERE hash = $1;

-- name: ExpireSessionRefreshTokens :exec
UPDATE auth.refresh_token
SET expires_at = @expires_at
WHERE session_id = @session_id AND revoked_at IS NULL
  AND expires_at > now();
