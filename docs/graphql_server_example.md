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
        []string{},
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





