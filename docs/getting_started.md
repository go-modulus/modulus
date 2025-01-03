## Getting Started

This is a guide to help you get started with the project. It will walk you through the steps to get the project up and running on your local machine.

### Installation
First, you need to install the Modulus CLI tool. You can do this by running the following command:

```bash
go install github.com/go-modulus/modulus/cmd/mtools@latest
```

Next, you need to initialize a new project. You can do this by running the following command:

```bash
mtools init --path=./testproj --name=testproj
```

If the `init` command runs without parameters it will prompt you to enter a name for your project. You can enter any name you like, but for this guide, we will use `testproj`.

### Adding Modules
Once you have initialized your project, you can add modules to it. Modules are reusable components that provide functionality to your project. To add a module, run the following command:

```bash
mtools module install --proj-path=./testproj -m "pgx"
```

or
    
```bash 
mtools module install --proj-path=./testproj
```

if you want to select the modules from the list.

Or even
    
```bash
cd testproj
mtools module install
```

if you want to install the modules in the current directory.

### Create the new module
To create a new module, run the following command:

```bash
mtools module create --proj-path=./testproj --silent --path=internal --package=example
```

It creates a new module in the `internal` directory with the `example` package. Also, the storage based on SQLc will be added to the module.
All initializations of the module will be done automatically:
* Adding NewModule to the `cmd/console/main.go`
* The module file with DI dependencies will be created in the `internal/example` directory.
* Also, the `internal/example/storage/migraion` and `internal/example/storage/query` directories are initialized with default files. Fill free to remove them and create the own ones.

### Migrate the database
First of all, run the local PostgreSQL on your machine or in docker.
Configure the database connection according the PG_* variables in the `.env` file of the project. 
Or fill these variables with your own values.


Next, you need to create the new migration. You are free to use any tool that is supported by [SQLc](https://docs.sqlc.dev/en/stable/howto/ddl.html#handling-sql-migrations), 
but we recommend using [dbmate](https://github.com/amacneil/dbmate) for this purpose. If you want to use dbmate, you are welcome to run the following command in mtools:

```bash
mtools db add --proj-path=./testproj --module=example --name=create_table
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

Then you need to run the migration:

```bash
mtools db migrate --proj-path=./testproj
```

You can also rollback the migration:

```bash
mtools db rollback --proj-path=./testproj
```


### Work with DB
To work with the database, you need to create a file with queries that you want to use in your project.

In our example create SQL queries for the new table in the `testproj/internal/example/storage/query` folder. 

We propose to name the file according to the name of a table, in our cease is `example.sql` and fill it with the SQL queries that work with the example table.

For example:

```sql
-- name: FindExamples :many
SELECT *
FROM example;
```

Read more about SQLc and formating its queries [here](https://docs.sqlc.dev/en/stable/tutorials/getting-started-postgresql.html).

After writing all the necessary queries, you need to update the SQLc configuration (`sqlc.yaml`) file if it was changed:

```bash
mtools db update-sqlc-config --proj-path=./testproj
```

Then you can generate the code for the queries:

```bash
mtools db generate --proj-path=./testproj
```

### Burn migrations into the binary
If you want to have all the migrations in your binary, you can burn them into it.
In this case mtools cannot help you, but you can add the `dbmate migrator` module to the `cmd/console/main.go` to have the migrate and rollback commands in the binary.

```bash
mtools module install --proj-path=./testproj -m "dbmate migrator"
```

Then you can run the following command to
```bash
go run cmd/console/main.go migrator migrate
```