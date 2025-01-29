-- migrate:up
CREATE SCHEMA IF NOT EXISTS blog;

CREATE TYPE blog.post_status AS ENUM ('draft', 'published', 'deleted');

CREATE TABLE blog.post
(
    id           uuid PRIMARY KEY,
    title        text      NOT NULL,
    preview      text      NOT NULL,
    content      text      NOT NULL,
    status       blog.post_status NOT NULL DEFAULT 'draft',
    created_at   timestamp NOT NULL DEFAULT now(),
    updated_at   timestamp NOT NULL DEFAULT now(),
    published_at timestamp,
    deleted_at   timestamp
);

-- migrate:down
DROP TABLE blog.post;
DROP TYPE blog.post_status;
DROP SCHEMA blog;
