# Testing

This section covers how to write and run tests for your application built on the Modulus framework.

## Writing tests
The framework is built on dependency injection, which requires some setup before you can write tests.

We recommend using a `main_test.go` file to configure the DI container for tests and initialize the required dependencies.

Let's write a test for the `CreatePost` function from the `blog` resolvers. We haven't implemented a separate action for the post creation, so let's test the resolver file.

The first step is to create a new file `main_test.go` in the `internal/blog/graphql` directory. It will run automatically before starting the tests.

```go
// package should be the same as the package of the test file
package graphql_test

import (
	"testing"

	"github.com/go-modulus/demo/internal/blog"
	"github.com/go-modulus/demo/internal/blog/graphql"
	"github.com/go-modulus/modulus/module"
	"github.com/go-modulus/modulus/test"
	"go.uber.org/fx"
)

func createMod() *module.Module {
	return blog.NewModule()
}

var (
	// all dependencies that you want to use in tests
	resolver *graphql.Resolver
)

func TestMain(m *testing.M) {
	test.LoadEnv()
	// create a new module where tested code is placed
	mod := createMod()

	test.TestMain(
		m,
		// add all necessary dependencies to the module
		module.BuildFx(mod),
		fx.Populate(
			// populate all dependencies that you want to use in tests
			&resolver,
		),
	)
}
```


Now we can write a test for the `CreatePost` function. We will test the resolver file, so we need to create a new test file `resolvers_test.go` in the `internal/blog/graphql` directory.

```go
package graphql_test

import (
	"blog/internal/graphql/model"
	"context"
	"github.com/go-modulus/modulus/auth"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestResolver_CreatePost(t *testing.T) {
	t.Parallel()
	t.Run("create post", func(t *testing.T) {
		t.Parallel()
		ctx := auth.WithPerformer(
			context.Background(),
			auth.Performer{
				ID:        uuid.Must(uuid.NewV6()),
			},
		)
		post, err := resolver.CreatePost(ctx, model.CreatePostInput{
			Title:   "Title",
			Content: "Content",
		})
		
		t.Log("When the post is created with valid input")
		t.Log("	Then the post should be created successfully")
		require.NoError(t, err)
		require.NotEqual(t, uuid.Nil, post.ID)
		require.Equal(t, "Title", post.Title)
		require.Equal(t, "Content", post.Content)
	})
}
```

While not required, we recommend naming tests as `StructName_MethodName` to make it clear what is being tested.
Also, you can use the `t.Run` function to describe the different test cases of the tested method.
Moreover, you can use the `t.Log` function to describe the test case as:

```text
Given the condition
When the action is performed
    Then the expectation A should be met
    And the expectation B should be met
```

Note that when testing the resolver directly, the context must include the current user ID. Use the `auth.WithPerformer` function to set it up, as shown above.
Business logic is usually placed in actions rather than resolvers. Actions receive the current user ID as an explicit input parameter, not from the context. When testing a resolver directly, however, the context must be prepared manually.

## Running tests
To run the tests, you need to execute the following command:

```bash
make test
```

The framework generates a `Makefile` in the project root with all the commands needed to run the project. The `test` target runs all tests.

Run this target and inspect the output.

It has to look like the next strings if the test is working:

```text
 resolvers_test.go:30: When the post is created with valid input
    resolvers_test.go:31:       Then the post should be created successfully
--- PASS: TestResolver_CreatePost (0.00s)
    --- PASS: TestResolver_CreatePost/create_post (0.02s)
PASS
ok      blog/internal/blog/graphql  0.276s
```

You may notice that by default tests run against the same database as the main application. You need to clean up the database after each test — either by running a regular query or by using fixtures generated with SQLc (covered in the next section).

It is also a good idea to run tests against a dedicated database. To do this:
1. Create a new env file `.env.test`.
2. Add `PGX_DSN=postgres://postgres:foobar@localhost:5432/blog_test?sslmode=disable`
3. Run migrations for the test environment
```bash
APP_ENV=test make db-migrate
```
4. Run tests
```bash
make test
```

## Fixtures

Fixtures provide predefined data for tests, letting you validate application behavior against a known, controlled dataset.

To use fixtures, first generate fixture builders with SQLc:

1. Edit the `internal/blog/storage/sqlc.tmpl.yaml` file and add the following code:

```yaml
sqlc-tmpl:
  options:
    fixture:
      overrides:
        *default-overrides
    ...    
  sql:
    ...
    codegen:
      - <<: *codegen-fixture
        options:
          <<: *codegen-fixture-options
          default_schema: "blog"
          package: "fixture"
          model_import: "your/package/path/of/storage"
```

2. Run the following command to generate the fixture builders:

```bash
make db-sqlc-generate
```
It will generate the `internal/blog/storage/fixture` directory with the post fixture builder.

3. We recommend creating a factory to initialize the builder with sensible default values.

```go
package fixture

import (
	"blog/internal/blog/storage"
	"github.com/gofrs/uuid"
	"gopkg.in/guregu/null.v4"
	"time"
)

type Factory struct {
	db storage.DBTX
}

func NewFactory(db storage.DBTX) *Factory {
	return &Factory{
		db: db,
	}
}

func (f *Factory) NewPostFixture() *PostFixture {
	id := uuid.Must(uuid.NewV6())
	return NewPostFixture(f.db, storage.Post{
		ID:          id,
		Title:       "Title " + id.String(),
		Preview:     "Preview " + id.String(),
		Content:     "Content " + id.String(),
		Status:      storage.PostStatusPublished,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		PublishedAt: null.TimeFrom(time.Now()),
		DeletedAt:   null.Time{},
		AuthorID:    uuid.Must(uuid.NewV6()),
	})
}
```
We recommend initializing fields with unique values. This avoids conflicts between tests and makes it safe to run them in parallel.

4. Add the factory to `main_test.go`. You could also add it to `module.go`, but keeping it in `main_test.go` is preferable to avoid bloating the DI container in production.

```go
package graphql_test

import (
	"testing"

	"github.com/go-modulus/demo/internal/blog"
	"github.com/go-modulus/demo/internal/blog/graphql"
	"github.com/go-modulus/demo/internal/blog/storage/fixture"
	"github.com/go-modulus/modulus/module"
	"github.com/go-modulus/modulus/test"
	"go.uber.org/fx"
)

func createMod() *module.Module {
	return blog.NewModule().
		// add the factory to the module's dependencies
		AddProviders(fixture.NewFactory)
}

var (
	// all dependencies that you want to use in tests
	resolver *graphql.Resolver
	// add a local variable of a factory to create fixtures in tests
	fixtures *fixture.Factory
)

func TestMain(m *testing.M) {
	test.LoadEnv()
	// create a new module where tested code is placed
	mod := createMod()

	test.TestMain(
		m,
		// add all necessary dependencies to the module
		module.BuildFx(mod),
		fx.Populate(
			// populate all dependencies that you want to use in tests
			&resolver,
			// populate fixtures to work with them in tests
			&fixtures,
		),
	)
}

```

5. Now you can use the factory to create fixtures in tests or cleanup the created data.

```go 
    post, err := resolver.CreatePost(
        ...
    )
    
    fixtures.NewPostFixture().ID(post.ID).Cleanup(t)
```
This registers the post for deletion after the test completes.

Run the test and inspect the database — it should be clean after the test, with only data that was not managed by the test remaining.

You can also use the fixture builder to verify that data was correctly persisted to the database.

Change the line
```go
fixtures.NewPostFixture().ID(post.ID).Cleanup(t)
```
with
```go
savedPost := fixtures.NewPostFixture().ID(post.ID).PullUpdates(t).Cleanup(t).GetEntity()
```

And add the following checks to the test:
```go
    t.Log("	And the post should be saved in the database")
    require.Equal(t, post.ID, savedPost.ID)
    require.Equal(t, post.Title, savedPost.Title)
    require.Equal(t, post.Content, savedPost.Content)
```

**That's it!** Now you can test your code with fixtures and clean up the database after each test.

