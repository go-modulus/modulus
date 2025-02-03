# Modulus Framework

Modulus is a web framework that allows you to create web applications with ease. It is designed to be simple and easy to use, and it is built on top of the Uber's [FX framework](https://github.com/uber-go/fx).

It is a framework to build modular server applications. It is designed to be simple and easy to use.
This project tends to answer the common developer questions:
- How to build a modular monolithic server application?
- How to decompose a codebase into modules to avoid increasing the code complexity?
- How to build a server application that can be easily extended and maintained?
- How to build a server application that can be easily tested?
- How to build a server application that can be easily deployed?

The first version of the project allows developers to create only two types of applications: cli app, and GraphQL web server.

In both cases, the application is composed of modules.
Each module is a directory that contains the module's code and configuration.
The module's code is divided to packages. The structure of packages can be different, but we recommend to use the following structure:
- `graphql` - contains the module's GraphQL resolvers, schemas and other GraphQL related code (for example directives). We recommend to use https://gqlgen.com/ implementation of GraphQL library, and all our standard modules and approaches in working with GraphQL are based on this library.
- `storage` - contains the module's storage code (for example database models, DAO objects, etc.). We recommend to use https://sqlc.dev/ library for working with DB.
    - `migration` - a list of SQL migrations that should be applied to the database. We recommend to use https://github.com/amacneil/dbmate tool for managing migrations.
    - `queries` - a set of files with SQL queries that are used in the module and which will be compiled to the Golang code by SQLc.
    - `fixture` -  a set of generated fixtures to fill the database with test data.
    - `dataloader` - a set of generated dataloaders aimed to prevent N+1 problem in GraphQL queries. We recommend to use https://github.com/graph-gophers/dataloader library.
- `cli` - contains the module's CLI commands.
- `action` - contains the module's business logic divided to separate structures with only one public method `Exec(ctx, request)`. Each action changes the state of the application and returns changed data.

The framework is based on the following awesome libraries:
- [Uber's FX framework](https://github.com/uber-go/fx) - a dependency injection framework.
- [SQLC](https://sqlc.dev/) - a library for generating Go code from SQL queries.
- [GQLGen](https://gqlgen.com/) - a library for generating Go code from GraphQL schemas.
- [Temporal](https://temporal.io/) - a workflow engine for building resilient, scalable, and distributed applications.


Read next:
- [Getting Started](getting_started.md) - to learn how to start using Modulus.
- [GraphQL Server Example](graphql_server_example.md) - to see a real example of a GraphQL server built with Modulus.