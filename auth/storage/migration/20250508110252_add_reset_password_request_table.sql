-- migrate:up
CREATE TYPE auth.reset_password_status AS ENUM ('active', 'expired', 'used');
CREATE TABLE auth.reset_password_request
(
    id           UUID                       NOT NULL PRIMARY KEY,
    account_id   UUID                       NOT NULL,
    status       auth.reset_password_status NOT NULL DEFAULT 'active',
    token        text               NOT NULL,
    last_send_at TIMESTAMP WITH TIME ZONE,
    used_at      TIMESTAMP WITH TIME ZONE,
    created_at   TIMESTAMP WITH TIME ZONE   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT reset_password_token_uniq UNIQUE (token)
);

-- migrate:down
DROP TABLE auth.reset_password_request;
DROP TYPE auth.reset_password_status;
