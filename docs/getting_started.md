# Getting Started

This is a guide to help you get started with the project. It will walk you through the steps to get the project up and running on your local machine.

## Installation
First, you need to install the Modulus CLI tool. You can do this by running the following command:

```bash
go install github.com/go-modulus/mtools/cmd/mtools@latest
```

Next, you need to initialize a new project. You can do this by running the following command:

```bash
mtools init --path=./testproj --name=testproj
```

If the `init` command is run without parameters, it will prompt you to enter a name for your project. You can enter any name you like, but for this guide we will use `testproj`.

## Adding Modules
Once you have initialized your project, you can add modules to it. Modules are reusable components that provide functionality to your project. To add a module, run the following command:

```bash
mtools module install --proj-path=./testproj -m "modulus/pgx"
```

or

```bash
mtools module install --proj-path=./testproj
```

if you want to select the modules from the list.

Or even just

```bash
cd testproj
mtools module install
```

if you want to install modules in the current directory.

By default, the module installer uses the registry of available modules in the [registry](https://github.com/go-modulus/registry) repository.
The file https://github.com/go-modulus/registry/blob/main/modules.json is used as the default module registry.
If you have your own manifest with custom modules, you can specify a path to a local or remote JSON registry file using the `--registry` flag. For a local manifest, use an absolute or relative path such as `./my-folder/manifest.json`.
For a remote manifest, pass a valid URL beginning with the `https://` prefix.

For example:

```bash
mtools module install --proj-path=./testproj --registry="./modules.json" -m "modulus/pgx"
```


### Create a New Module
To create a new module, run the following command:

```bash
mtools module create --proj-path=./testproj --silent --path=internal --package=example
```
or just
```bash
cd testproj
mtools module create
```

This creates a new module in the `internal` directory with the `example` package. Storage based on SQLc is also added to the module.
All module initialization is done automatically:
* `NewModule` is added to `cmd/console/main.go`.
* The module file with DI dependencies is created in the `internal/example` directory.
* The `internal/example/storage/migration` and `internal/example/storage/query` directories are initialized with default files. Feel free to remove them and create your own.

## Migrate the Database
First, run a local PostgreSQL instance on your machine or in Docker.
Configure the database connection using the `PG_*` variables in the `.env` file of the project, or set these variables to your own values.


Next, you need to create a new migration. You are free to use any tool supported by [SQLc](https://docs.sqlc.dev/en/stable/howto/ddl.html#handling-sql-migrations),
but we recommend using [dbmate](https://github.com/amacneil/dbmate) for this purpose. To use dbmate via mtools, run the following command:

```bash
mtools db add --proj-path=./testproj --module=example --name=create_table
```
or just
```bash
cd testproj
mtools db add
```

Find the created migration in the `testproj/internal/example/storage/migration` directory and fill it with the necessary SQL code.

For example:
```sql
-- migrate:up
CREATE TABLE example (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

-- migrate:down
DROP TABLE example;
```

Then run the migration:

```bash
mtools db migrate --proj-path=./testproj
```
or just
```bash
cd testproj
mtools db migrate
```

You can also roll back the migration:

```bash
mtools db rollback --proj-path=./testproj
```


### Work with the Database
To work with the database, create a file containing the queries you want to use in your project.

In our example, we create SQL queries for the new table in the `testproj/internal/example/storage/query` folder.

We recommend naming the file after the table — in our case `example.sql` — and filling it with the SQL queries that operate on the example table.

For example:

```sql
-- name: FindExamples :many
SELECT *
FROM example;
```

Read more about SQLc and formatting its queries [here](https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html).

After writing all the necessary queries, update the SQLc configuration (`sqlc.yaml`) file if it has changed:

```bash
mtools db update-sqlc-config --proj-path=./testproj
```

Then generate the code for the queries:

```bash
mtools db generate --proj-path=./testproj
```

### Embed Migrations in the Binary
If you want to bundle all migrations into your binary, you can do so by adding the `modulus/pgx/migrator` module to `cmd/console/main.go`. This gives you `migrate` and `rollback` commands directly in the binary.

```bash
mtools module install --proj-path=./testproj -m "modulus/pgx/migrator"
```

Then run the following command to migrate the database using the project binary:
```bash
cd testproj
go run cmd/console/main.go migrator migrate
```

### Create a New Command
To create a new CLI command, run the following command:

```bash
mtools module add-cli --proj-path=./testproj --module=example --name=hello-world
```
or configure everything interactively
```bash
cd testproj
mtools module add-cli
```

This creates a new CLI command in the `internal/example/cli/hello_world.go` file.
All constructors are also added to the `internal/example/module.go` file,
so you can start writing business logic in the new CLI command without any boilerplate configuration.
Feel free to modify the command to suit your needs.
Afterwards, you can run the command:

```bash
cd testproj
go run cmd/console/main.go hello-world
```

### Add an HTTP Endpoint
So far we have worked only with CLI commands, but now we are going to add an HTTP endpoint to our project.
To do this, install the `http` module, which adds the ability to run an HTTP server in your application.

```bash
mtools module install --proj-path=./testproj -m "http"
```

Then create a new HTTP handler:

```bash
mtools module add-json-api --proj-path=./testproj --module=example --uri=/hello-world --name=HelloWorld --method=GET --silent
```
or configure everything interactively
```bash
cd testproj
mtools module add-json-api
```

This creates a new HTTP handler in the `internal/example/api/hello_world.go` file.
Feel free to modify the handler to suit your needs.
Afterwards, you can run the HTTP server:

```bash
cd testproj
make install
./bin/console serve
```

If everything is working, you will see output like the following in the console:

```bash
2025-01-13T14:24:58+02:00       DEBUG   registering route       {"app": "modulus", "path": "/hello-world", "component": "http", "method": "GET"}
2025-01-13T14:24:58+02:00       INFO    http server is starting {"app": "modulus", "component": "http"}
2025-01-13T14:24:59+02:00       INFO    http server has started {"app": "modulus", "component": "http", "address": "localhost:8001"}
```

Now open your browser and navigate to `http://localhost:8001/hello-world` to see the handler's response.


## Install the GraphQL Module
We use the `modulus/graphql` module to add GraphQL support to our project. It is based on the https://gqlgen.com library, a powerful tool for building a GraphQL server from GraphQL schemas.

To install the `modulus/graphql` module, run the following command:
```bash
mtools module install --proj-path=./testproj -m "modulus/graphql"
```
or select it interactively
```bash
cd testproj
mtools module install
```

After that, you can test the GraphQL server by running the following command:

```bash
cd testproj
make install
./bin/console serve
```

If everything is working, you will see output like the following in the console:

```bash
2025-01-20T15:03:35+02:00       INFO    http server has started {"app": "modulus", "component": "http", "address": "localhost:8001"}
```

Now open your browser and navigate to `http://localhost:8001/playground` to see the GraphQL playground.

Write and run the following query in the playground:

```graphql
{
  ping
}
```

You should see the following result:

```json
{
  "data": {
    "ping": "pong"
  }
}
```