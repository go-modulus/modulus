-- name: CreateIdentity :one
insert into "auth"."identity"
    (id, identity, user_id, "data")
values (@id::uuid, @identity::text, @user_id::uuid, @data::jsonb)
RETURNING *;

-- name: DeleteIdentity :exec
delete from "auth"."identity"
where id = @id;

-- name: FindIdentity :one
select *
from "auth"."identity"
where identity = @identity::text;

-- name: FindUserIdentities :many
select *
from "auth"."identity"
where user_id = @user_id::uuid;