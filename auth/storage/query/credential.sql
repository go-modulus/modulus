-- name: CreateCredential :one
INSERT INTO "auth"."credential"
    (account_id, type, hash, expired_at)
VALUES (@account_id::uuid, @type::text, @hash::text, @expired_at)
RETURNING *;

-- name: FindLastCredential :one
SELECT *
FROM "auth"."credential"
WHERE account_id = @account_id::uuid
ORDER BY created_at DESC;

-- name: FindLastCredentialOfType :one
SELECT *
FROM "auth"."credential"
WHERE account_id = @account_id::uuid
AND type = @type::text
ORDER BY created_at DESC;

-- name: FindAllCredentialsOfType :many
SELECT *
FROM "auth"."credential"
WHERE type = @type::text
ORDER BY created_at DESC;

-- name: RemoveCredentialsOfAccount :exec
DELETE FROM "auth"."credential"
WHERE account_id = @account_id::uuid;