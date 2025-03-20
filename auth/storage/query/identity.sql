-- name: CreateIdentity :one
insert into "auth"."identity"
    (id, identity, account_id, "data", "type")
values (@id::uuid, @identity::text, @account_id::uuid, @data::jsonb, @type::text)
RETURNING *;

-- name: RemoveIdentity :exec
delete from "auth"."identity"
where id = @id;

-- name: FindIdentity :one
select *
from "auth"."identity"
where identity = @identity::text;

-- name: FindAccountIdentities :many
select *
from "auth"."identity"
where account_id = @account_id::uuid;

-- name: FindIdentityById :one
select *
from "auth"."identity"
where id = @id::uuid;

-- name: BlockIdentity :exec
update "auth"."identity"
set status = 'blocked'::auth.identity_status
where id = @id::uuid;

-- name: BlockIdentitiesOfAccount :exec
update "auth"."identity"
set status = 'blocked'::auth.identity_status
where account_id = @account_id::uuid;

-- name: ActivateIdentity :exec
update "auth"."identity"
set status = 'active'::auth.identity_status
where id = @id::uuid;

-- name: RequestIdentityVerification :exec
update "auth"."identity"
set status = 'not-verified'::auth.identity_status
where id = @id::uuid;

-- name: RemoveIdentitiesOfAccount :exec
delete from "auth"."identity"
where account_id = @account_id::uuid;