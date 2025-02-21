-- migrate:up
ALTER TABLE blog.post
    ADD COLUMN author_id uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';

-- migrate:down
ALTER TABLE blog.post
    DROP COLUMN author_id;

