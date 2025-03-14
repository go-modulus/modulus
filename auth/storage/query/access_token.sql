-- name: CreateAccessToken :one
INSERT INTO auth.access_token (hash, identity_id, session_id, user_id, roles, data, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: GetAccessTokenByHash :one
SELECT *
FROM auth.access_token
WHERE hash = $1;

-- name: RevokeAccessToken :exec
UPDATE auth.access_token
SET revoked_at = now()
WHERE hash = $1 AND revoked_at IS NULL;

-- name: RevokeSessionAccessTokens :exec
UPDATE auth.access_token
SET revoked_at = now()
WHERE session_id = $1 AND revoked_at IS NULL;

-- name: RevokeUserAccessTokens :exec
UPDATE auth.access_token
SET revoked_at = now()
WHERE user_id = $1 AND revoked_at IS NULL;

-- name: GetUserNotRevokedSessionIds :many
SELECT session_id
FROM auth.access_token
WHERE revoked_at IS NULL
  AND user_id = $1;

-- name: ExpireSessionAccessTokens :exec
UPDATE auth.access_token
SET expires_at = @expires_at
WHERE session_id = @session_id AND revoked_at IS NULL
  AND expires_at > now();