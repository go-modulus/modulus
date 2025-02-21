# Real GraphQL Server Example

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

## Requirements

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

## Blog Module SQL
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

## Blog Module GraphQL

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


## Resolvers Implementation

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

There are no data in the database yet. Let's create a new post. We need to implement the `createPost` mutation resolver.

Add the following code to the `internal/blog/graphql/resolvers.go` file:

```go   
import(
    "github.com/go-modulus/modulus/validator"
    validation "github.com/go-ozzo/ozzo-validation/v4"
    "github.com/gofrs/uuid"
)

func (r *Resolver) CreatePost(ctx context.Context, input model.CreatePostInput) (storage.Post, error) {
    // validate input using Ozzo validation
    err := validation.ValidateStructWithContext(
        ctx,
        &input,
        validation.Field(
            &input.Title,
            validation.Required.Error("Title is required"),
        ),
        validation.Field(
            &input.Content,
            validation.Required.Error("Content is required"),
        ),
    )
    if err != nil {
        return storage.Post{}, validator.NewErrInvalidInputFromOzzo(ctx, err)
    }
    
    preview := input.Content
    if len(input.Content) > 100 {
        preview = input.Content[0:100]
    }
    
    return r.blogDb.CreatePost(
        ctx, storage.CreatePostParams{
            ID:      uuid.Must(uuid.NewV6()),
            Title:   input.Title,
            Preview: preview,
            Content: input.Content,
        },
    )
}

```

The `CreatePost` function validates the input using the `Ozzo` validation library. If the input is invalid, the function returns an error.
The function creates a new post with the `draft` status and the `preview` field that contains the first 100 characters of the `content` field.
Try it in playground:

```graphql

mutation {
  createPost(input:{title:"aaa", content:"bbb"}){id, title, content}
}

```

It will return the new post with the `id`, `title`, and `content` fields.

```json
{
  "data": {
    "createPost": {
      "id": "f3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b",
      "title": "aaa",
      "content": "bbb"
    }
  }
}
```

But the post is still in the `draft` status. Let's implement the `publishPost` mutation resolver.

Add the following code to the `internal/blog/graphql/resolvers.go` file:

```go
func (r *Resolver) PublishPost(ctx context.Context, id uuid.UUID) (storage.Post, error) {
    return r.blogDb.PublishPost(ctx, id)
}
```

The `PublishPost` function publishes the post with the provided `id`. Try it in playground:

```graphql
mutation {
  publishPost(id:"f3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"){id, title, content, status}
}
```

Run the `posts` query. You will see errors like these: 
```
{
  "errors": [
    {
      "message": "Something went wrong on our side (RID: )",
      "path": [
        "posts",
        0,
        "publishedAt"
      ],
      "extensions": {
        "code": "panic: not implemented: PublishedAt - publishedAt"
      }
    },
    {
      "message": "Something went wrong on our side (RID: )",
      "path": [
        "posts",
        0,
        "deletedAt"
      ],
      "extensions": {
        "code": "panic: not implemented: DeletedAt - deletedAt"
      }
    }
```

It is because the Go types of fields in the DB models and in Graphql models differs.
For each field with different type the resolver is added by the `gqlgen` generator.
Let's fill this resolver with the type conversion.

Open the `internal/graphql/resolver/schema.resolvers.go` file and change the following code:

```go
func (r *postResolver) PublishedAt(ctx context.Context, obj *storage.Post) (*time.Time, error) {
	panic(fmt.Errorf("not implemented: PublishedAt - publishedAt"))
}
```

to

```go
func (r *postResolver) PublishedAt(ctx context.Context, obj *storage.Post) (*time.Time, error) {
    if obj.PublishedAt.Valid {
        return &obj.PublishedAt.Time, nil
    }
    return nil, nil
}
```

Also, you can see the `deletedAt` field in the result, but it is a little bit system field. Let's hide it from the result with `createdAt` and `updatedAt`.
Open the `internal/blog/blog/storage/sqlc.tmpl.yaml` file and add fields:

```yaml
      codegen:
        - <<: *codegen-graphql
          options:
            ...
+            exclude:
+              - "Post.deletedAt"
+              - "Post.updatedAt"
+              - "Post.createdAt"
```

Run the SQLc and GraphQL generations again:

```bash
    make db-sqlc-generate
    make graphql-generate
```

Call the query

```graphql
{
    posts {
        id
        title
        preview
        content
        status
        publishedAt
    }
}
```

and you will see the list of posts without errors.


## Adding users
According to the requirements, we need to add users and create posts for them. Let's create the `user` table and queries for it.

```shell 
     make module-create
```

Enter the `user` module name and the `user` schema name. Also, chose all the default values.
The result view of all selected options are there: 
![create user module](./img/create_user_module.png)

Create the new migration in the `internal/user/storage/migration` directory:

```shell
    make db-add
    unlink internal/user/storage/migration/default_schema.sql
    unlink internal/user/storage/query/default_query.sql
``` 

The selected options are there:
![create user table](./img/create_user_table.png)

Add the following code to the created file of the migration:

```sql
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
```

Add users queries to the `internal/user/storage/query/user.sql` file:

```sql
-- name: RegisterUser :one
INSERT INTO "user"."user" (id, email, name)
VALUES (@id::uuid, @email::text, @name::text)
RETURNING *;

-- name: FindUserByEmail :one
SELECT * FROM "user"."user"
WHERE email = @email::text;
```

Run the checking the new SQL migration. It will appy the migration, rollback it and apply again.
It is a good practice to check a new migration for rollback and apply it again.

```shell
    make db-check-migration
``` 
Run the SQLc generation to make the queries available in the code:

```shell
    make db-sqlc-generate
```

Change the `internal/user/storage/sqlc.tmpl.yaml` file to generate GraphQL types by SQLc:

```yaml
sqlc-tmpl:
  version: "2"
  options:
    graphql:
      overrides:
        *default-overrides
    ...
  sql:  
    codegen:
    - <<: *codegen-graphql
      options:
        <<: *codegen-graphql-options
        default_schema: "user"
        package: "blog/internal/user/storage"
```

Run the generation to create the schema:

```shell
    make db-sqlc-generate
```

Add the `user` resolvers to the `internal/user/graphql/resolvers.go` file:

```go
package graphql

type Resolver struct {
	
}

func NewResolver() *Resolver {
	return &Resolver{}
}
```

Add the `userResolver` resolver to the `internal/graphql/resolver/resolver.go` file:

```go
type Resolver struct {
    // Place all dependencies here
+   userResolver *userGraphql.Resolver
}

func NewResolver(
	...
    userResolver *userGraphql.Resolver,
) *Resolver {
    return &Resolver{
		...
        userResolver: userResolver,
    }
}
```

Add the `user.graphql` schema to the `internal/user/graphql` directory:

```graphql
extend type Mutation {
    registerUser(input: RegisterUserInput!): User!
} 

input RegisterUserInput @goModel(model: "blog/internal/user/action.RegisterUserInput") {
    email: String!
    password: String!
    name: String!
}
```

As you can see we linked the `RegisterUserInput` to the `RegisterUserInput` struct in the `blog/internal/user/action` package.
Let's create an action for the user module. Create the `internal/user/action` directory and the `register_user.go` file in it.

```shell
    mkdir internal/user/action
    touch internal/user/action/register_user.go
```

Action is a struct with one public method `Execute` that returns the result of the action.

```go
package action

import (
	"blog/internal/user/storage"
	"braces.dev/errtrace"
	"context"
	"errors"
	"github.com/go-modulus/modulus/errors/erruser"
	"github.com/go-modulus/modulus/validator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrUserAlreadyExists = erruser.New("user already exists", "User already exists. Please login or use another email.")

type RegisterUserInput struct {
	Email    string
	Password string
	Name     string
}

func (i *RegisterUserInput) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(
		i,
		validation.Field(
			&i.Email,
			validation.Required.Error("Email is required"),
			is.Email.Error("Email is not valid"),
		),
		validation.Field(
			&i.Password,
			validation.Required.Error("Password is required"),
			validation.Length(6, 20).Error("Password must be between 6 and 20 characters"),
		),
		validation.Field(
			&i.Name,
			validation.Required.Error("Name is required"),
			is.Alpha.Error("Name must contain only letters"),
		),
	)

	if err != nil {
		return validator.NewErrInvalidInputFromOzzo(ctx, err)
	}

	return nil
}

type RegisterUser struct {
	userDb *storage.Queries
}

func NewRegisterUser(userDb *storage.Queries) *RegisterUser {
	return &RegisterUser{userDb: userDb}
}

func (r *RegisterUser) Execute(ctx context.Context, input RegisterUserInput) (storage.User, error) {
	err := input.Validate(context.Background())
	if err != nil {
		return storage.User{}, err
	}

	_, err = r.userDb.FindUserByEmail(ctx, input.Email)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return storage.User{}, errtrace.Wrap(err)
		}
	} else {
		return storage.User{}, ErrUserAlreadyExists
	}
	user, err := r.userDb.RegisterUser(
		ctx, storage.RegisterUserParams{
			ID: uuid.Must(uuid.NewV6()),
			Email: input.Email,
			Name:  input.Name,
		},
	)
	if err != nil {
		return storage.User{}, errtrace.Wrap(err)
	}
	return user, nil
}
```

Add the `RegisterUser` action to the module providers in the `internal/user/module.go` file:

```go
import (
    "blog/internal/user/action"
    "blog/internal/user/graphql"
)
func NewModule() *module.Module {
    return module.NewModule("user").
        ...
        AddProviders(
        ...
        action.NewRegisterUser,
		graphql.NewResolver,
    )
}
```

Add a link to the `RegisterUser` action in the `internal/user/graphql/resolvers.go` file:

```go
package graphql

import (
	"blog/internal/user/action"
	"blog/internal/user/storage"
	"context"
)

type Resolver struct {
	register *action.RegisterUser
}

func NewResolver(
	register *action.RegisterUser,
) *Resolver {
	return &Resolver{
		register: register,
	}
}

func (r *Resolver) RegisterUser(ctx context.Context, input action.RegisterUserInput) (storage.User, error) {
	return r.register.Execute(ctx, input)
}
```

Generate the GraphQL resolvers:

```shell
    make graphql-generate
```

In the generated `internal/graphql/resolver/user.resolvers.go` file, add the following code:

```go
func (r *mutationResolver) RegisterUser(ctx context.Context, input action.RegisterUserInput) (storage.User, error) {
	return r.userResolver.RegisterUser(ctx, input)
}
```


Now we can register a new user. Try it in playground:

```graphql
mutation {
  registerUser(input:{email:"test@test.com", password:"123456", name:"Test"}){id, email, name}
}
```



## Authenticate User
We have the `registerUser` mutation to create a new user. Now we need to authenticate the user.
To get the basement for our authentication we can use the `auth` module of the Modulus framework.
Let's install the `auth` module with the following command:

```shell
    make module-install
```
And select `modulus auth` module from the list of available modules.

After installing the `auth` module, we need to update the schema of our DB with the new tables for the `auth` module.
Migrations have been created in the `internal/auth/storage/migration`, so we need to run them:

```shell
    make db-migrate
```

Now let's make an identity for the further authentication at the `RegisterUser` action.

Add to the `internal/user/action/register_user.go` file:

```go
type RegisterUser struct {
	userDb *storage.Queries
	passwordAuth *auth.PasswordAuthenticator
}

func NewRegisterUser(
	userDb *storage.Queries,
    passwordAuth *auth.PasswordAuthenticator,
) *RegisterUser {
	return &RegisterUser{
		userDb: userDb,
		passwordAuth: passwordAuth,
	}
}

func (r *RegisterUser) Execute(ctx context.Context, input RegisterUserInput) (storage.User, error) {
	...
	_, err = r.passwordAuth.Register(
        ctx,
        input.Email,
        input.Password,
        user.ID,
		// the authenticated user role that will be used in the future
        []string{"user"},
        nil, 
    )
    if err != nil {
        return storage.User{}, errtrace.Wrap(err)
    }
	return user, nil
}

```

Make a login mutation in the `internal/user/graphql/user.graphql` file:

```graphql

extend type Mutation {
    ...
    loginUser(input: LoginUserInput!): TokenPair!
}


input LoginUserInput @goModel(model: "blog/internal/user/action.LoginUserInput") {
    email: String!
    password: String!
}

type TokenPair @goModel(model: "blog/internal/user/action.TokenPair") {
    accessToken: String!
    refreshToken: String!
}

```

Generate resolvers and add the `loginUser` resolver to the `internal/user/graphql/resolvers.go` file. Link added resolver with generated resolver in the `internal/graphql/resolver/user.resolvers.go` file.
This steps the same as we did for the `registerUser` mutation.

Also, don't forget to make the `LoginUser` action in the `internal/user/action/login_user.go` file.
Call its constructor in the `internal/user/module.go` file.
And call Execute method in the `internal/user/graphql/resolvers.go` file.

The `LoginUser` action should look like this:

```go
package action

import (
	"braces.dev/errtrace"
	"context"
	"github.com/go-modulus/modulus/auth"
	"github.com/go-modulus/modulus/validator"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type LoginUserInput struct {
	Email    string
	Password string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (i *LoginUserInput) Validate(ctx context.Context) error {
	err := validation.ValidateStruct(
		i,
		validation.Field(
			&i.Email,
			validation.Required.Error("Email is required"),
			is.Email.Error("Email is not valid"),
		),
		validation.Field(
			&i.Password,
			validation.Required.Error("Password is required"),
			validation.Length(6, 20).Error("Password must be between 6 and 20 characters"),
		),
	)

	if err != nil {
		return validator.NewErrInvalidInputFromOzzo(ctx, err)
	}

	return nil
}

type LoginUser struct {
	passwordAuth   *auth.PasswordAuthenticator
	plainTokenAuth *auth.PlainTokenAuthenticator
}

func NewLoginUser(
	passwordAuth *auth.PasswordAuthenticator,
	tokenAuth *auth.PlainTokenAuthenticator,
) *LoginUser {
	return &LoginUser{
		passwordAuth:   passwordAuth,
		plainTokenAuth: tokenAuth,
	}
}

// Execute performs the login action by email and password.
// Returns a token pair of the access and refresh tokens if the login is successful.
// Errors:
// * github.com/go-modulus/modulus/auth.ErrIdentityIsBlocked - if the identity is blocked.
// * github.com/go-modulus/modulus/auth.ErrInvalidPassword - if the password is invalid.
// * Any error from the IdentityRepository.Get method (e.g. github.com/go-modulus/modulus/auth/repository.ErrIdentityNotFound).
// * Any error from the CredentialRepository.GetLast method (e.g. github.com/go-modulus/modulus/auth/repository.ErrCredentialNotFound).
func (l *LoginUser) Execute(ctx context.Context, input LoginUserInput) (TokenPair, error) {
	// Authenticate the user with the given email and password.
	performer, err := l.passwordAuth.Authenticate(ctx, input.Email, input.Password)
	if err != nil {
		return TokenPair{}, errtrace.Wrap(err)
	}

	// Issue a new pair of access and refresh tokens.
	pair, err := l.plainTokenAuth.IssueTokens(ctx, performer.IdentityID, nil)
	if err != nil {
		return TokenPair{}, errtrace.Wrap(err)
	}

	return TokenPair{
		AccessToken:  pair.AccessToken.Token.String,
		RefreshToken: pair.RefreshToken.Token.String,
	}, nil
}
```

Try it in playground:

```graphql
mutation {
  registerUser(input:{email:"test3@test.com", password:"123456", name:"Test"}){id, email, name}
}
``` 
to register a new user.

And then try to login:

```graphql
mutation {
  loginUser(input:{email:"test3@test.com", password:"123456"}){
    accessToken, 
    refreshToken
  }
}
```

You will get the `accessToken` and `refreshToken` in the response.

```
{
  "data": {
    "loginUser": {
      "accessToken": "vDl1BirAEcQv917FjSs1dTfYa/y1gySu",
      "refreshToken": "79GWIAkZOKNQeQAjrIOsQPDsm7jubwkW"
    }
  }
}
```

In this example we will not use the `refreshToken` but it is a good practice to use it in the real project.


## Protect Queries and Mutations

First of all, let's protect the `createPost` mutation to allow only authenticated users to create posts.

We need to use the Auth middleware for reading the tokens from headers and checking the access token. Also, this middleware will add the `Performer` to the context.

Add the following code to the `/cmd/console/main.go`:

```go
    modules := []*module.Module{
    ...
        http.NewModule().AddProviders(
			func(authMd *auth.Middleware) *http.Pipeline {
				return &http.Pipeline{
					Middlewares: []http.Middleware {
						authMd.HttpMiddleware(),
					},
				}
			},
		),
	...
    }
```

It setups the pipeline of middlewares for the HTTP server. The `authMd.HttpMiddleware()` adds the `Auth` middleware to the pipeline.
You can see the usage of the `HttpMiddleware()` method instead of `Middleware`. It is because the `Middleware` method has our own signature and is not compatible with the Chi router.
The method `HttpMiddleware()` wraps our vision of the middleware to the standard middleware representation.

Regenerate the graphql resolvers:

```shell
    make graphql-generate
```

Now we need a guard directive for securing the GraphQL queries.

Add `AuthGuard` directive to the `internal/graphql/resolver/resolver.go` file:

```go
import (
    "blog/internal/auth/graphql"
)
func (r Resolver) GetDirectives() generated.DirectiveRoot {
	return generated.DirectiveRoot{
		AuthGuard: graphql.AuthGuard,
	}
}
```

This directive has been created by the auth module when we installed it. It has a default logic of comparing roles in Performer with roles passed to the directive.
Feel free to change the logic if you need it.
In our example this simple logic is enough.

Now protect a query `createPost()` with the `AuthGuard` directive in the `internal/blog/graphql/blog.graphql` file:

```graphql
extend type Mutation {
    createPost(input: CreatePostInput!): Post! @authGuard(allowedRoles: ["user"])
}
```

We have added the `@authGuard(allowedRoles: ["user"])` directive to the `createPost` mutation. It means that only authenticated users with the role `user` can create posts.
In a case if admin should have the ability to create posts, you can add the `admin` role to the `allowedRoles` list.

Regenerate the GraphQL resolvers:

```shell
    make graphql-generate
```

Now try to create a post without authentication:

```graphql
mutation {
  createPost(input:{title:"aaa", content:"bbb"}){id, title, content}
}
```

You will get an error message like this:

```json
{
  "errors": [
    {
      "message": "Please authenticate to get access to this resource",
      "path": [
        "createPost"
      ],
      "extensions": {
        "code": "unauthenticated"
      }
    }
  ],
  "data": null
}
```

Now try to put the `accessToken` to the `Authorization` header and run the mutation again.

In Playground, click on the `Headers` tab and add the following header:

```json
{"Authorization": "Bearer <your access token obtained from the 'loginUser()' mutation>"}
```

Now try to create a post again. You will get the new error message:

```json
{
  "errors": [
    {
      "message": "You are not authorized to perform this action",
      "path": [
        "createPost"
      ],
      "extensions": {
        "code": "unauthorized"
      }
    }
  ],
  "data": null
}
```

It is because the `accessToken` doesn't contain the `user` role. In your application you obviously will create a user management sub-system, but in this example we just add necessary data to the database manually.

Write `{user}` to the field `roles` in the `auth.identity` table for the user you want to authenticate.

Login again to get the access token with updated roles and try to create a post again.

Now everything should work fine.

Protect also the `publishPost` and `deletePost` mutations. 


## Add author to post

We need to add the author to the post. Let's create a new migration to add the `author_id` field to the `post` table.

```shell
    make db-add
```

Enter the following code to the created file of the migration:

```sql  
-- migrate:up
ALTER TABLE blog.post
    ADD COLUMN author_id uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';

-- migrate:down
ALTER TABLE blog.post
    DROP COLUMN author_id;
```

Check the new migration:

```shell
    make db-check-migration
```

Update the `post.sql` file in the `internal/blog/storage/query` directory with the following content:

```sql
-- name: CreatePost :one
INSERT INTO blog.post (id, author_id, title, preview, content)
VALUES (@id::uuid, @author_id::uuid, @title::text, @preview::text, @content::text)
RETURNING *;
```

Here we added the `author_id` field to the `CreatePost` query.

Run the SQLc generation to make the queries available in the code:

```shell
    make db-sqlc-generate
```

Edit the `internal/blog/graphql/resolvers.go` file and add the `AuthorID` field to the `CreatePostParams`:

```go
    authorId := auth.GetPerformerID(ctx)
	return r.blogDb.CreatePost(
		ctx, storage.CreatePostParams{
			...
			AuthorID: authorId,
		},
	)
```

Try to create a post and get convinced that the `author_id` field is filled with the `id` of the authenticated user.

If you see on the generated GraphQL `Post` type you see that the `authorId: Uuid!` field is added. But we need to add the `author` field of type `User` to the `Post` type.

Add the following code to the `internal/blog/graphql/blog.graphql` file:

```graphql
extend type Post {
    author: User!
}
```

And avoid generating the `authorId` field adding the line to the `internal/blog/storage/sqlc.tmpl.yaml`

```yaml
   exclude:
   - "Post.deletedAt"
   ...
   - "Post.authorId"
```

Regenerate SQLc and GraphQL:

```shell
    make db-sqlc-generate
    make graphql-generate
```

After that we get the new resolver for the `author` field in the `internal/graphql/resolver/blog.resolvers.go` file:

```go
// Author is the resolver for the author field.
func (r *postResolver) Author(ctx context.Context, obj *storage.Post) (storage1.User, error) {
	panic(fmt.Errorf("not implemented: Author - author"))
}
```

We will implement it in the next chapter describing the dataloaders concept.

Now we want to show both all published posts and drafts of the current user. Let's change the `posts` query to return the list of posts of the current user.

Let's change the `FindPosts` query in the `internal/blog/storage/query/post.sql` file:

```sql
-- name: FindPosts :many
SELECT *
FROM blog.post
WHERE status = 'published'
   or (status = 'draft' and author_id = @author_id::uuid)
ORDER BY published_at DESC;
```

Generate the SQLc and change the `Posts` resolver in the `internal/blog/graphql/resolvers.go` file:

```go
func (r *Resolver) Posts(ctx context.Context) ([]storage.Post, error) {
    authorId := auth.GetPerformerID(ctx)
    return r.blogDb.FindPosts(ctx, authorId)
}
```

Try to call a list of posts in the playground:

```graphql
{
    posts {
        id
        title
        preview
        content
        status
        publishedAt
        author {
            id
            email
            name
        }
    }
}
```

with the `accessToken` in the `Authorization` header. You will see the list of errors instead of posts of the current user.

It is because we have not implemented the `Author` resolver yet. We will do it in the next chapter.


## Dataloaders

Dataloaders are a powerful tool for reducing the number of queries to the database. They allow you to batch and cache the requests to the database.
We use the https://github.com/graph-gophers/dataloader tool to make the dataloaders in our project.

You are free to use any other dataloader library or implement your own dataloaders.
But we recommend generate the dataloaders with the `sqlc` generator using the plugin https://github.com/debugger84/sqlc-dataloader.

Let's add the dataloaders to the project.
Edit your `/internal/user/storage/sqlc.tmpl.yaml` file adding the dataloader plugin:

```yaml
sqlc-tmpl:
  version: "2"
  options:
    ...
    dataloader:
      overrides:
        *default-overrides
    ...
  sql:
    - schema: "migration"
     ...
      codegen:
        - <<: *codegen-dataloader
          options:
            <<: *codegen-dataloader-options
            default_schema: "user"  
            model_import: "blog/internal/user/storage"
            cache:
              - table: "user.user"
                type: "lru"
                ttl: "1m"
                size: 100
```

Read more about the dataloader plugin options in the Readme of the plugin [repository](https://github.com/debugger84/sqlc-dataloader).

Run the SQLc generation:

```shell
    make db-sqlc-generate
```

It creates the dataloaders for the `user` module. Dataloaders are generated with unresolved dependencies so we need to call:

```shell
    go get github.com/debugger84/sqlc-dataloader
```

The generator creates the `internal/user/storage/dataloader` directory with the dataloaders. It also creates the `internal/user/storage/dataloader/loader_factory.go` that have to be used in your code as an entrypoint to the loaders.

Add the dataloaders to the `internal/user/module.go` file:

```go
import (
    "blog/internal/user/storage/dataloader"
)

func NewModule() *module.Module {
    return module.NewModule("user").
        ...
        AddProviders(
            ...
            dataloader.NewLoaderFactory,
        )
}
```

Also add the dependency to the dataloader factory to the `internal/graphql/resolver/resolver.go` file:

```go
import (
    userDataloader "blog/internal/user/storage/dataloader"
)

type Resolver struct {
    ...
	userLoaderFactory *userDataloader.LoaderFactory
}

func NewResolver(
    ...
    userLoaderFactory *userDataloader.LoaderFactory,
) *Resolver {
    return &Resolver{
        ...
        userLoaderFactory: userLoaderFactory,
    }
}
```

Load an author in the `internalgraphql/resolver/blog.resolvers.go` file:

```go
func (r *postResolver) Author(ctx context.Context, obj *storage.Post) (storage1.User, error) {
	return r.userLoaderFactory.UserLoader().Load(ctx, obj.AuthorID)
}
```

Rerun the server and try to get a list of posts again. You will see the list of posts with the author data.

```graphql
{
    posts {
        id
        title
        preview
        content
        status
        publishedAt
        author {
            id
            email
            name
        }
    }
}
```