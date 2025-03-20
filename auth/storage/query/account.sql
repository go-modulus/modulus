-- name: RegisterAccount :one
INSERT INTO auth.account (id)
VALUES ($1) RETURNING *;

-- name: FindAccount :one
SELECT *
FROM auth.account
WHERE id = $1;

-- name: BlockAccount :exec
UPDATE auth.account
SET status = 'blocked'::auth.account_status
WHERE id = $1;

-- name: UnblockAccount :exec
UPDATE auth.account
SET status = 'active'::auth.account_status
WHERE id = $1;


-- name: AddRoles :exec
update "auth"."account"
set roles = array(select distinct unnest(roles || @roles::text[]))
where id = @id::uuid;

-- name: RemoveRoles :exec
update "auth"."account"
set roles = array(select distinct unnest(roles) except select distinct unnest(@roles::text[]))
where id = @id::uuid;

-- name: RemoveAccount :exec
DELETE FROM auth.account
WHERE id = $1;