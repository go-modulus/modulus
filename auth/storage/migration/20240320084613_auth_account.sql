-- migrate:up

CREATE TYPE auth.account_status AS ENUM (
    'active',
    'blocked'
    );

ALTER TABLE auth.identity
    RENAME COLUMN user_id TO account_id;
ALTER TABLE auth.identity
    DROP COLUMN roles;
ALTER TABLE auth.identity
    ADD COLUMN type text NOT NULL DEFAULT 'not-set';

COMMENT ON COLUMN auth.identity.type IS 'Type of the identity (eg. email, phone, google-auth, etc.).';

ALTER TYPE auth.identity_status ADD VALUE IF NOT EXISTS 'not-verified';

CREATE TABLE auth.account
(
    id         uuid PRIMARY KEY,
    status     auth.account_status NOT NULL DEFAULT 'active'::auth.account_status,
    roles      text[]              NOT NULL DEFAULT '{}',
    data       jsonb,
    updated_at timestamptz         NOT NULL DEFAULT NOW(),
    created_at timestamptz         NOT NULL DEFAULT NOW()
);

ALTER TABLE auth.access_token
    RENAME COLUMN user_id TO account_id;
ALTER TABLE auth.session
    RENAME COLUMN user_id TO account_id;

ALTER TABLE auth.credential
    RENAME COLUMN identity_id TO account_id;

-- migrate:down
ALTER TABLE auth.identity
    RENAME COLUMN account_id TO user_id;
ALTER TABLE auth.identity
    ADD COLUMN roles text[] NOT NULL DEFAULT '{}';
ALTER TABLE auth.identity
    DROP COLUMN type;

ALTER TABLE auth.access_token
    RENAME COLUMN account_id TO user_id;
ALTER TABLE auth.session
    RENAME COLUMN account_id TO user_id;

ALTER TABLE auth.credential
    RENAME COLUMN account_id TO identity_id;

DROP TABLE auth.account;
DROP TYPE auth.account_status;