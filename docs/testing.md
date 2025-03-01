# Testing

In this section, we will discuss how to test your code. We will show you how to run the tests and how to write the tests for your code.

## Writing tests
First of all, you need to understand that our framework is based on the dependency injection principle, so it is not so easy to write tests for the code.

We propose to use `main_test.go` file to describe DI for the tests and init necessary dependencies.

Let's write a test for the `CreatePost` function from the `blog` resolvers. We haven't implemented a separate action for the post creation, so let's test the resolver file.

The first step is to create a new file `main_test.go` in the `internal/blog/graphql` directory. It will run automatically before starting the tests.

```go
// package should be the same as the package of the tested file
package graphql_test

import (
	"blog/internal/blog"
	"blog/internal/blog/graphql"
	"github.com/go-modulus/modulus/module"
	"github.com/go-modulus/modulus/test"
	"go.uber.org/fx"
	"testing"
)

var (
	// all dependencies that you want to use in tests
	resolver   *graphql.Resolver
)

func TestMain(m *testing.M) {
	// put the path to the root of the project. And load the environment variables from there
	test.LoadEnv("../../..")
	// create a new module where tested code is placed
	mod := blog.NewModule()

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

It is not necessary, but we recommend you to name a test as `StructName_MethodName` to make it easier to understand what is being tested.
Also, you can use the `t.Run` function to describe the different test cases of the tested method.
Moreover, you can use the `t.Log` function to describe the test case as:

```text
Given the condition
When the action is performed
    Then the expectation A should be met
    And the expectation B should be met
```

I want to note that due to the testing of the resolver file, we need to prepare context with the current user ID. We use the `auth.WithPerformer` function to do this.
Usually, the business logic is placed in actions, but not in resolvers. And actions gets the current user ID from the input parameter instead of the context. But in a case of testing the resolver we need to prepare the context manually.

## Running tests
To run the tests, you need to execute the following command:

```bash
make test
```

Our framework creates a make file with all necessary commands to run the project. You can find it in the root of the project. The `test` command runs all tests in the project.

It looks like the test is working:

```text
 resolvers_test.go:30: When the post is created with valid input
    resolvers_test.go:31:       Then the post should be created successfully
--- PASS: TestResolver_CreatePost (0.00s)
    --- PASS: TestResolver_CreatePost/create_post (0.02s)
PASS
ok      blog/internal/blog/graphql  0.276s
```

But look at your database. It uses the same database as the main application. So, you need to clean up the database after each test. You can do it calling a regular query to DB, or use our fixtures that are generated with SQLc.

By the way, it is a good idea to run test over the separate DB. You are able to do this by:
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

Fixtures are a set of predefined data that is used as input for tests. They are used to test the application with a known set of data. Fixtures can be used to test the application with different data sets.

To make fixtures, you need to generate fixture builders with SQLc.

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
          model_import: "blog/internal/blog/storage"
```

2. Run the following command to generate the fixture builders:

```bash
make db-sqlc-generate
```
It will generate the `internal/blog/storage/fixture` directory with the post fixture builder.

3. It is a good idea to create a factory to initialize the builder with the default values.

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
We recommend you to initialize fields with unique values. It will help you to avoid conflicts between tests and simplifies running the test in a parallel mode.

4. Add the factory to the `main_test.go` file. You can add it to the `module.go` file, but we recommend you to add it to the `main_test.go` file to avoid swelling of the container.

```go
var (
	resolver *graphql.Resolver
	// add a local variable of a factory to create fixtures in tests
	fixtures *fixture.Factory
)

func TestMain(m *testing.M) {
	test.LoadEnv("../../..")
	mod := blog.NewModule().
	// add the factory to the module's dependencies
		AddProviders(fixture.NewFactory)

	test.TestMain(
		m,
		module.BuildFx(mod),
		fx.Populate(
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
It marks that the post with ID obtained from the `CreatePost` function should be deleted after the test running.

Run the test and check the database. It should be cleaned up after the test running (only old data that is not managed by test left there).


Let's change the test to get convinced that the data is stored in the database.
It also can be achieved via the fixture builder.

Change the line
```go
fixtures.NewPostFixture().ID(post.ID).Cleanup(t)
```
with
```go
savedPost := fixtures.NewPostFixture().ID(post.ID).PullUpdates(t).Cleanup(t).GetEntity()
```

Amd add the following checks to the test:
```go
    t.Log("	And the post should be saved in the database")
    require.Equal(t, post.ID, savedPost.ID)
    require.Equal(t, post.Title, savedPost.Title)
    require.Equal(t, post.Content, savedPost.Content)
```

**That's it!** Now you can test your code with fixtures and clean up the database after each test.

