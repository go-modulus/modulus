# Modulus Framework - CLAUDE.md

## Project Overview
Modulus is a Go framework for building modular server applications, supporting both CLI applications and GraphQL web servers. It's designed for creating maintainable, testable, and easily extensible monolithic applications with a modular architecture.

## Framework Architecture
- **Modular Design**: Applications are composed of independent modules with clear dependencies
- **Dependency Injection**: Uses Uber's FX framework for dependency injection and lifecycle management
- **CLI & GraphQL Support**: Primary focus on CLI applications and GraphQL web servers
- **Database Integration**: Built-in support for PostgreSQL with SQLc for type-safe queries

## Core Modules Available

### Essential Infrastructure
- **CLI Module** (`cli`): Creates CLI applications using `urfave/cli/v2`
- **HTTP Module** (`http`): HTTP server based on Chi router with middleware support
- **GraphQL Module** (`graphql`): GraphQL server using gqlgen with playground
- **Logger Module** (`logger`): Structured logging with slog and zap backend

### Database & Storage
- **PGX Module** (`db/pgx`): PostgreSQL driver and connection pooling
- **Migrator Module** (`db/migrator`): Database migrations using DBMate
- **Embedded Module** (`db/embedded`): Embedded PostgreSQL for testing

### Authentication & Security
- **Auth Module** (`auth`): Token-based authentication with sessions
- **Email Provider** (`auth/providers/email`): Email-based authentication
- **Captcha Module** (`captcha`): CAPTCHA integration for forms

### Advanced Features
- **Temporal Module** (`temporal`): Workflow and activity management
- **Translation Module** (`translation`): Internationalization support
- **Error Handling** (`errors`): Comprehensive error management system

## Development Tools
- **mtools CLI**: Main development tool for project management
  - `mtools init`: Initialize new projects
  - `mtools module install`: Install framework modules
  - `mtools module create`: Create custom modules
  - `mtools db`: Database operations (migrate, rollback, generate)

## Project Structure
```
project/
├── cmd/                    # Application entry points
│   └── console/           # CLI application entry
├── internal/              # Private application code
│   ├── auth/             # Authentication module (if installed)
│   ├── graphql/          # GraphQL module (if installed)
│   └── [custom-modules]/ # Your custom modules
├── mk/                   # Makefile includes
├── bin/                  # Compiled binaries
├── modules.json          # Installed modules manifest
├── Makefile             # Development commands
└── .env files           # Environment configurations
```

## Module Structure (Recommended)
```
module/
├── module.go            # Module definition and providers
├── graphql/            # GraphQL resolvers and schemas
├── storage/            # Database layer
│   ├── migration/      # SQL migrations
│   ├── query/         # SQLc query files
│   └── fixture/       # Test data fixtures
├── action/            # Business logic actions
└── cli/               # CLI commands
```

## Configuration & Environment
- Environment-based configuration using `go-envconfig`
- `.env` files for different environments (dev, test, prod)
- Module-specific configuration with automatic env variable binding

## Testing & Development
- Built-in test helpers and fixtures
- Database testing with embedded PostgreSQL
- Snapshot testing support
- Make targets for common operations:
  - `make test`: Run all tests
  - `make analyze`: Static code analysis
  - `make db-migrate`: Run database migrations

## Key Technologies
- **Go 1.24**: Latest Go version
- **FX**: Dependency injection framework
- **Chi**: HTTP router
- **gqlgen**: GraphQL implementation
- **SQLc**: Type-safe SQL generation
- **DBMate**: Database migration tool
- **Temporal**: Workflow engine (optional)

## Development Workflow
1. Use `mtools init` to create new project
2. Install required modules with `mtools module install`
3. Create custom modules with `mtools module create`
4. Use provided Makefile targets for development tasks
5. Follow modular architecture patterns

## Build & Deployment
- Standard Go build process
- Docker support available
- Environment-specific configurations
- Database migrations handled via CLI tools

## Current Development Status
Based on git status, the project is actively developing an email authentication provider with:
- Email-based user registration and login
- Password reset functionality
- CAPTCHA integration
- GraphQL API endpoints
- Database migrations and storage layer

## Commands for AI Assistant
- **Build**: `make install` (builds mtools binary)
- **Test**: `make test` (runs all tests)
- **Lint**: `make analyze` (static analysis)
- **Database**: `make db-migrate` (run migrations)
- **Generate**: `make db-sqlc-generate` (generate SQLc code)

## Module Installation Examples
```bash
# Install database module
mtools module install -m "pgx"

# Install GraphQL module  
mtools module install -m "gqlgen"

# Install authentication
mtools module install -m "auth"
```

This framework emphasizes modularity, type safety, and developer productivity while maintaining the simplicity of a monolithic deployment model.