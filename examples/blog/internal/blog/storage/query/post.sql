-- name: CreatePost :one
INSERT INTO blog.post (id, author_id, title, preview, content)
VALUES (@id::uuid, @author_id::uuid, @title::text, @preview::text, @content::text)
RETURNING *;

-- name: FindPost :one
SELECT *
FROM blog.post
WHERE id = @id::uuid;

-- name: FindPosts :many
SELECT *
FROM blog.post
WHERE status = 'published'
or (status = 'draft' and author_id = @author_id::uuid)
ORDER BY published_at DESC;

-- name: PublishPost :one
UPDATE blog.post
SET status       = 'published',
    published_at = now()
WHERE status = 'draft'
  AND id = @id::uuid
RETURNING *;
