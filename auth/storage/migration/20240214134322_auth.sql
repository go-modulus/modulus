-- migrate:up
CREATE SCHEMA IF NOT EXISTS auth;
CREATE TYPE auth.identity_status AS ENUM (
    'active',
    'blocked'
    );

CREATE TABLE auth.identity
(
    id         uuid PRIMARY KEY,
    identity   text                 NOT NULL UNIQUE,
    user_id    uuid                 NOT NULL,
    status     auth.identity_status NOT NULL DEFAULT 'active'::auth.identity_status,
    data       jsonb,
    roles      text[]               NOT NULL DEFAULT '{}',
    updated_at timestamptz          NOT NULL DEFAULT NOW(),
    created_at timestamptz          NOT NULL DEFAULT NOW()
);

CREATE INDEX identity_user_id_idx ON auth.identity (user_id);

CREATE TABLE auth.refresh_token
(
    hash        text PRIMARY KEY,
    session_id  uuid        NOT NULL,
    identity_id uuid        NOT NULL,
    revoked_at  timestamptz,
    expires_at  timestamptz NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX refresh_token_session_id_idx ON auth.refresh_token (session_id);

-- Let's denormalize the data to get extra fast access in one query for all necessary data
CREATE TABLE auth.access_token
(
    hash        text PRIMARY KEY,
    identity_id uuid        NOT NULL,
    session_id  uuid        NOT NULL,
    user_id     uuid        NOT NULL,
    roles       text[]      NOT NULL,
    data        jsonb,
    revoked_at  timestamptz,
    expires_at  timestamptz NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX access_token_identity_id_idx ON auth.access_token (identity_id);
CREATE INDEX access_token_session_id_idx ON auth.access_token (session_id);

CREATE TABLE "auth".session
(
    id          uuid PRIMARY KEY,
    user_id     uuid        NOT NULL,
    identity_id uuid        NOT NULL,
    data        jsonb,
    expires_at  timestamptz NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE auth.credential
(
    hash        text PRIMARY KEY,
    identity_id uuid        NOT NULL,
    type        text        NOT NULL,
    expired_at  timestamptz          DEFAULT NOW(),
    created_at  timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX credential_identity_id_idx ON auth.credential (identity_id);

-- migrate:down
DROP TABLE auth.identity;
DROP TABLE auth.refresh_token;
DROP TABLE auth.access_token;
DROP TABLE auth.credential;
DROP TABLE auth.session;

DROP TYPE auth.identity_status;
DROP SCHEMA auth CASCADE;