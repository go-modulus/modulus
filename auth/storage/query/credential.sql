-- name: CreateCredential :one
INSERT INTO "auth"."credential"
    (identity_id, type, hash, expired_at)
VALUES (@identity_id::uuid, @type::text, @hash::text, @expired_at)
RETURNING *;

-- name: FindLastCredential :one
SELECT *
FROM "auth"."credential"
WHERE identity_id = @identity_id::uuid
ORDER BY created_at DESC;

-- name: FindLastCredentialOfType :one
SELECT *
FROM "auth"."credential"
WHERE identity_id = @identity_id::uuid
AND type = @type::text
ORDER BY created_at DESC;

-- name: FindAllCredentialsOfType :many
SELECT *
FROM "auth"."credential"
WHERE type = @type::text
ORDER BY created_at DESC;