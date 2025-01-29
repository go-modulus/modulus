-- name: CreatePost :one
INSERT INTO blog.post (id, title, preview, content)
VALUES (@id::uuid, @title::text, @preview::text, @content::text);

-- name: FindPost :one
SELECT *
FROM blog.post
WHERE id = @id::uuid;

-- name: FindPosts :many
SELECT *
FROM blog.post
WHERE status = 'published'
ORDER BY published_at DESC;

-- name: PublishPost :one
UPDATE blog.post
SET status       = 'published',
    published_at = now()
WHERE status = 'draft'
  AND id = @id::uuid;
