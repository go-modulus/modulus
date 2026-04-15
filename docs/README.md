# Modulus Framework

Modulus is a web framework that allows you to create web applications with ease. It is designed to be simple and easy to use, and is built on top of Uber's [FX framework](https://github.com/uber-go/fx).

It is a framework for building modular server applications.
This project aims to answer common developer questions:
- How do I build a modular monolithic server application?
- How do I decompose a codebase into modules to manage code complexity?
- How do I build a server application that is easy to extend and maintain?
- How do I build a server application that is easy to test?
- How do I build a server application that is easy to deploy?

The first version of the project allows developers to create two types of applications: a CLI app and a GraphQL web server.

In both cases, the application is composed of modules.
Each module is a directory that contains the module's code and configuration.
The module's code is divided into packages. The package structure can vary, but we recommend the following:
- `graphql` - contains the module's GraphQL resolvers, schemas, and other GraphQL-related code (for example, directives). We recommend using the https://gqlgen.com/ GraphQL library; all our standard modules and patterns are based on it.
- `storage` - contains the module's storage code (for example, database models, DAO objects, etc.). We recommend using the https://sqlc.dev/ library for database access.
    - `migration` - a list of SQL migrations to be applied to the database. We recommend using https://github.com/amacneil/dbmate for managing migrations.
    - `queries` - SQL query files used by the module, which are compiled to Go code by SQLc.
    - `fixture` - generated fixtures for populating the database with test data.
    - `dataloader` - generated dataloaders to prevent the N+1 problem in GraphQL queries. We recommend using the https://github.com/graph-gophers/dataloader library.
- `cli` - contains the module's CLI commands.
- `action` - contains the module's business logic, divided into separate structs each with a single public method `Exec(ctx, request)`. Each action mutates application state and returns the updated data.

The framework is based on the following awesome libraries:
- [Uber's FX framework](https://github.com/uber-go/fx) - a dependency injection framework.
- [SQLC](https://sqlc.dev/) - a library for generating Go code from SQL queries.
- [GQLGen](https://gqlgen.com/) - a library for generating Go code from GraphQL schemas.
- [Temporal](https://temporal.io/) - a workflow engine for building resilient, scalable, and distributed applications.


Read next:
- [Getting Started](getting_started.md) - to learn how to start using Modulus.
- [GraphQL Server Example](graphql_server_example.md) - to see a real example of a GraphQL server built with Modulus.