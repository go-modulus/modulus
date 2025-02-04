-- name: RegisterUser :one
INSERT INTO "user"."user" (id, email, name)
VALUES (@id::uuid, @email::text, @name::text)
RETURNING *;

-- name: FindUserByEmail :one
SELECT * FROM "user"."user"
WHERE email = @email::text;