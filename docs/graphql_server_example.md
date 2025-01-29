## Real GraphQL Server Example

Let's create a real-world example that provides a simple GraphQL API for the non-existent Front-end.
The requirements for the project will be described later. Now, let's say that we want to create an API for the blog.
It is a simple API, but it is enough to show how to create a GraphQL server using Modulus framework and the `gqlgen` module.
Previously we created a server that has only one query `ping` that returns the string `pong`.

Read the [Getting Started](./getting_started.md) guide to create a new project with the base set of modules.

If you don't want to read this and follow all the steps from the article, please use the following commands to create a new project:

```bash
	go install github.com/go-modulus/modulus/cmd/mtools@latest
	mkdir blog
    cd blog
	mtools init --name=blog
	mtools module install -m "pgx, dbmate migrator, chi http, gqlgen"
	mtools module create --silent --path=internal --package=blog
```

All next calls to `mtools` will be executed in the `blog` directory.

### Requirements

Every project starts with requirements. Let's define the requirements for our blog API:
1. There are 2 roles in the system: `Admin` and `User`.
2. The `Admin` can create, update, and delete any posts in the system. Also, the `Admin` can see all posts, including unpublished ones.
3. The `User` can create, update, and delete only his posts and see all published posts.
4. The system shows the list of posts with the following fields: `ID`, `Title`, `Preview`, `Author`, `Status`, `PublishedAt`.
5. The system allows filtering posts by `Status` and sorting by `PublishedAt`.
6. The system allows getting the full post by `ID` providing the additional field `Content`.
7. To work with a system the users should register themselves in the system.
8. The system should provide the `login` mutation to authenticate the user.
9. The system should provide the `me` query to get the current user.
10. The system should provide the `logout` mutation to log out the user.
11. The system should provide the `refresh` mutation to refresh the user's token.
12. The `Admin` can see the list of all users in the system.
13. The `Admin` can change the role of the user.

### Prepare SQL of the blog  module
First of all, we need to create the blog module. Let's create the `blog` module with the following command:

```bash
    mtools module create --silent --path=internal --package=blog
```

After that, we need to define the database schema. Let's use the migrate module to create the first migration:

```bash
     mtools db add --module=blog --name=create_schema
```

The command will create the `<timestamp>_create_schema.sql` file in the `internal/blog/storage/migration` directory. 
For the first iteration let's define the schema for the `post` table only. We will add the schema for the `user` table later.

```sql
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
```

After creating the schema, we may delete the default migration:

```bash
    unlink internal/blog/storage/migration/default_schema.sql
```

Cool! Now we have the schema for the `post` table. Let's create the queries for the `post` table that will be used in our code.
We need:
1. The `CreatePost` SQL-query to create a new post. The post will be in the status `draft` by default.
2. The `FindPost` query to get the post by `ID`.
3. The `FindPosts` query to get the list of posts with filtering and sorting.
4. The `PublishPost` query to publish the post.

Create the `post.sql` file in the `internal/blog/storage/query` directory with the following content:

```sql
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
```

Remove the default query. We don't need it anymore:

```bash
    unlink internal/blog/storage/query/default_query.sql
```

One more thing, by default, sqlc is configured to work with the public schema. But in our case we use the `blog` schema.
Let's configure `internal/blog/storage/sqlc.tmpl.yaml` to use the `blog` schema:

```yaml
- default_schema: "public"
+ default_schema: "blog"
```

Now we have to run generation to make the queries available in the code:

```bash
    make db-sqlc-update
```
