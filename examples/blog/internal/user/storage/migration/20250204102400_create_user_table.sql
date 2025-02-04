-- migrate:up

CREATE SCHEMA IF NOT EXISTS "user";

CREATE TABLE "user"."user" (
    id uuid PRIMARY KEY,
    email text NOT NULL unique CHECK (email ~* '^.+@.+\..+$'),
    name text NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

-- migrate:down
DROP TABLE "user"."user";
DROP SCHEMA "user";
