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
VALUES (@id::uuid, @title::text, @preview::text, @content::text)
RETURNING *;

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
  AND id = @id::uuid
RETURNING *;
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
    make db-sqlc-generate
```

It will generate the `internal/blog/storage/post.sql.go` file with the `CreatePost`, `FindPost`, `FindPosts`, and `PublishPost` functions.
Also, it will generate the `internal/blog/storage/models.go` file with the `Post` struct and the `PostStatus` enum.

Now we need to configure the database connection. Open .env file and change the following line to your local database connection:

```env
PGX_DSN=postgres://postgres:foobar@localhost:5432/test?sslmode=disable
```

If you don't want to use DSN as a configuration feel free to use the separate environment variables for the database connection. But don't forget to comment the `PG_DSN` variable.

```env
DB_NAME=test
HOST=localhost
PASSWORD=foobar
# Use this variable to set the DSN for the PGX connection. It overwrites the other PG_* variables.
#PGX_DSN=postgres://postgres:foobar@localhost:5432/test?sslmode=disable
PORT=5432
SSL_MODE=disable
USER=postgres
```

After doing this, run migrations to create the necessary db with all tables:

```bash
    make db-migrate
```

### Create the graphql resolvers

Now we need to create the resolvers for the blog module. Let's create the `internal/blog/graphql` directory and the `resolvers.go` file in it.
The resolvers will be used to handle the GraphQL queries and mutations.

```shell
    mkdir internal/blog/graphql
    touch internal/blog/graphql/resolvers.go
```

Let's define the resolvers structure:

```go
package graphql

type Resolver struct {
	
}
```

Now we need to inject it to the `internal/graphql/resolver/resolver.go` file:

```go
type Resolver struct {
	// Place all dependencies here
+	blogResolver *blogGraphql.Resolver
}
func NewResolver(
+   blogResolver *blogGraphql.Resolver,
) *Resolver {
    return &Resolver{
+       blogResolver: blogResolver,
    }
}
```

We can create a schema in `schema.graphql` file in the `internal/blog/graphql` directory defining the Post type and queries and mutations for the Post.
But it is so boring. Let's use SQLc plugin to generate the schema for us.

Add an anchor for the `codegen-graphql` and `codegen-graphql-options` to the `internal/blog/storage/sqlc.tmpl.yaml` file:
```yaml
sqlc-tmpl:
  options:
    ...
+    graphql:
+      overrides:
+        *default-overrides
  sql:
    ...
      codegen:
+       - <<: *codegen-graphql
+          options:
+            <<: *codegen-graphql-options
+            package: "blog/internal/blog/storage"  
+            default_schema: "blog"
        - <<: *codegen-golang
```

Now we need to run the generation to create the schema:

```bash
    make db-sqlc-generate
```

It will generate the `internal/blog/graphql/schema.graphql` file with the `Post` type:

```graphql
enum PostStatus  @goModel(model: "blog/internal/blog/storage.PostStatus") {
    draft
    published
    deleted
}


type Post @goModel(model: "blog/internal/blog/storage.Post") {
    id: Uuid!
    title: String!
    preview: String!
    content: String!
    status: PostStatus!
    createdAt: Time!
    updatedAt: Time!
    publishedAt: Time
    deletedAt: Time
}
```

Also, we need queries and mutation for the `Post` type. Let's add the following code to the `internal/blog/graphql/blog.graphql` file:

```graphql
extend type Query {
    post(id: ID!): Post
    posts: [Post!]!
}

extend type Mutation {
    createPost(input: CreatePostInput!): Post!
    publishPost(id: Uuid!): Post!
    deletePost(id: Uuid!): Boolean!
}

input CreatePostInput {
    title: String!
    content: String!
}
```

Run the generation to create the resolvers:

```bash
    make graphql-generate
```

It will generate the `internal/graphql/resolver/blog.resolvers.go` file with blog resolvers.

```go
// CreatePost is the resolver for the createPost field.
func (r *mutationResolver) CreatePost(ctx context.Context, input model.CreatePostInput) (storage.Post, error) {
	panic(fmt.Errorf("not implemented: CreatePost - createPost"))
}

// PublishPost is the resolver for the publishPost field.
func (r *mutationResolver) PublishPost(ctx context.Context, id uuid.UUID) (storage.Post, error) {
	panic(fmt.Errorf("not implemented: PublishPost - publishPost"))
}

// DeletePost is the resolver for the deletePost field.
func (r *mutationResolver) DeletePost(ctx context.Context, id uuid.UUID) (bool, error) {
	panic(fmt.Errorf("not implemented: DeletePost - deletePost"))
}

// Post is the resolver for the post field.
func (r *queryResolver) Post(ctx context.Context, id string) (*storage.Post, error) {
	panic(fmt.Errorf("not implemented: Post - post"))
}

// Posts is the resolver for the posts field.
func (r *queryResolver) Posts(ctx context.Context) ([]storage.Post, error) {
	panic(fmt.Errorf("not implemented: Posts - posts"))
}
```

Copy the resolvers to the `internal/blog/graphql/resolvers.go` file:

```go
package graphql

import (
	"blog/internal/blog/storage"
	"blog/internal/graphql/model"
	"context"
	"fmt"
	"github.com/gofrs/uuid"
)

type Resolver struct {
}

// CreatePost is the resolver for the createPost field.
func (r *Resolver) CreatePost(ctx context.Context, input model.CreatePostInput) (storage.Post, error) {
	panic(fmt.Errorf("not implemented: CreatePost - createPost"))
}

// PublishPost is the resolver for the publishPost field.
func (r *Resolver) PublishPost(ctx context.Context, id uuid.UUID) (storage.Post, error) {
	panic(fmt.Errorf("not implemented: PublishPost - publishPost"))
}

// DeletePost is the resolver for the deletePost field.
func (r *Resolver) DeletePost(ctx context.Context, id uuid.UUID) (bool, error) {
	panic(fmt.Errorf("not implemented: DeletePost - deletePost"))
}

// Post is the resolver for the post field.
func (r *Resolver) Post(ctx context.Context, id string) (*storage.Post, error) {
	panic(fmt.Errorf("not implemented: Post - post"))
}

// Posts is the resolver for the posts field.
func (r *Resolver) Posts(ctx context.Context) ([]storage.Post, error) {
	panic(fmt.Errorf("not implemented: Posts - posts"))
}
```

And call these resolvers from the `internal/graphql/resolver/blog.resolvers.go` file:

```go

// CreatePost is the resolver for the createPost field.
func (r *mutationResolver) CreatePost(ctx context.Context, input model.CreatePostInput) (storage.Post, error) {
	return r.blogResolver.CreatePost(ctx, input)
}

// PublishPost is the resolver for the publishPost field.
func (r *mutationResolver) PublishPost(ctx context.Context, id uuid.UUID) (storage.Post, error) {
	return r.blogResolver.PublishPost(ctx, id)
}

// DeletePost is the resolver for the deletePost field.
func (r *mutationResolver) DeletePost(ctx context.Context, id uuid.UUID) (bool, error) {
	return r.blogResolver.DeletePost(ctx, id)
}

// Post is the resolver for the post field.
func (r *queryResolver) Post(ctx context.Context, id string) (*storage.Post, error) {
	return r.blogResolver.Post(ctx, id)
}

// Posts is the resolver for the posts field.
func (r *queryResolver) Posts(ctx context.Context) ([]storage.Post, error) {
	return r.blogResolver.Posts(ctx)
}
```

After that run the server:

```bash
    make install
    ./bin/console serve
```

Open the `http://localhost:8080/playground` in the browser and try to run the following query:

```graphql
{
    posts {
        id
        title
        preview
        content
        status
        createdAt
        updatedAt
        publishedAt
        deletedAt
    }
}
```

You will have an error message like this: `Something went wrong on our side`. It is because we didn't implement the resolvers yet.


### Implement the resolvers

Go to the `internal/blog/graphql/resolvers.go` file and add dependency to the DB to the `Resolver` struct:

```go
type Resolver struct {
	blogDb *storage.Queries
}

func NewResolver(blogDb *storage.Queries) *Resolver {
	return &Resolver{blogDb: blogDb}
}
```

Add the `NewResolver` constructor to the module providers in the `internal/blog/module.go` file:

```go
func NewModule() *module.Module {
    return module.NewModule("blog").
		...
		AddProviders(
+			graphql.NewResolver,
```

Now we can implement the resolvers. Let's start with the `posts` query resolver located in the `internal/blog/graphql/resolvers.go` file:

```go
func (r *Resolver) Posts(ctx context.Context) ([]storage.Post, error) {
	return r.blogDb.FindPosts(ctx)
}
```

The `FindPosts` function returns the list of posts with the status `published` and sorts them by the `published_at` field.

Build server program again and rerun it. Try to get posts again. You will see the empty list of posts in the response without any errors.

The result will be like this:

```json
{
  "data": {
    "posts": []
  }
}
```
