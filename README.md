# Project Probation: Hotel Booking System

A Go REST API robust backend for a Hotel Booking System following Hexagonal Architecture principles. It features user authentication (register/login), hotel and room search, real-time availability checking, booking management, and simulated payments. The system is engineered to guarantee strict data consistency and entirely prevent double bookings under any concurrent scenarios, utilizing PostgreSQL persistence, Sentry error reporting, and comprehensive testing with testcontainers.

---

## Table of Contents

- [Architecture](#architecture)
- [Development](#development)
- [Configuration](#configuration)
- [Running the Server](#running-the-server)
- [Database Migrations](#database-migrations)
- [Testing](#testing)
- [Linting](#linting)
- [Docker](#docker)
- [Project Structure](#project-structure)

---

## Architecture

This project follows **Hexagonal Architecture** with clear separation of concerns:

![Hexagonal Architecture Diagram](docs/arch.png)

Below is a Mermaid diagram illustrating the architecture and dependencies:

```mermaid
graph TB
    subgraph "Infrastructure Layer (Adapters)"
        HTTP[HTTP Server<br/>Echo Framework<br/>httpserver/]
        DB[PostgreSQL<br/>GORM<br/>postgres/]
        PKG[Cross-cutting<br/>Config, Sentry<br/>pkg/]
    end
    
    subgraph Hexagon["(the hexagon)"]
        subgraph "Application"
            UC[Use Cases<br/>Business Logic<br/>contact/usecase.go]
        end
        
        subgraph "Domain"
            ENT[Contact Entity<br/>Validation<br/>contact/contact.go]
            INTF[Interfaces<br/>Service<br/>Repository]
        end
    end
    
    HTTP -->|depends on| UC
    DB -->|implements| INTF
    UC -->|uses| INTF
    UC -->|validates| ENT
    
    style ENT fill:#e1f5ff
    style INTF fill:#e1f5ff
    style UC fill:#fff4e1
    style HTTP fill:#ffe1e1
    style DB fill:#ffe1e1
    style PKG fill:#ffe1e1
```

**Layers:**
- **Domain Layer** ([`user/`](user/)): User entity, validation logic, and interface definitions (`Service`, `Repository`)
- **Application Layer** ([`user/usecase.go`](user/usecase.go), [`auth/usecase.go`](auth/usecase.go)): Business logic implementation (user management + authentication)
- **Infrastructure Layer**: Adapters for external dependencies
  - [`httpserver/`](httpserver/) - Echo web framework HTTP handlers
  - [`postgres/`](postgres/) - GORM database implementation
  - [`pkg/`](pkg/) - Cross-cutting concerns (config, sentry, logging, hashing, jwt, oauth)

**Key Principles:**
- Dependencies point inward: Infrastructure depends on domain, never the reverse
- Domain defines interfaces; infrastructure implements them
- Dependency injection throughout the stack
- Comprehensive error handling with custom error codes ([`errs/`](errs/))

---

## Development

### Prerequisites

- Go 1.24.0 or higher
- Docker & Docker Compose (for local PostgreSQL)
- Air (for hot reload): `go install github.com/air-verse/air@latest`
- sql-migrate (for migrations): `go install github.com/rubenv/sql-migrate/...@latest`
- golangci-lint (for linting)

### Quick Start

1. **Initialize the project:**
   ```shell
   go mod download
   ```

2. **Start local database:**
   ```shell
   make local-db
   ```
   This starts PostgreSQL in Docker on port `33062`.

3. **Run database migrations:**
   ```shell
   make db/migrate
   ```

4. **Start the server with hot reload:**
   ```shell
   make run
   ```
   The server will start on port `8088` (default) and auto-reload on code changes.

5. **Run tests:**
   ```shell
   make test
   ```

### Development Workflow

- **Hot reload** is configured via [`.air.toml`](.air.toml)
  - Watches `*.go` files (excludes `*_test.go`)
  - Builds to `tmp/main` from `cmd/httpserver/`
  - Auto-restarts on file changes
- Use `make lint` to validate code quality
- Use `make format` to run `gofmt -w .`
- Use `make swag` to generate Swagger docs from annotations
- Integration tests use `testcontainers` for isolated PostgreSQL instances

---

## Configuration

Configuration is managed through environment variables loaded via [`.env`](.env) file (see [`pkg/config/config.go`](pkg/config/config.go)).

### Environment Variables

Create a `.env` file in the project root:

```env
APP_ENV=local
PORT=8088
SENTRY_DSN=your-sentry-dsn-here
ALLOW_ORIGINS=*

DB_HOST=localhost
DB_USER=root
DB_PASS=123456
DB_PORT=33062
DB_NAME=teqlocal
ENABLE_SSL=false
```

### Configuration Loading

The application uses `envconfig` to load environment variables:
- Automatically loads `.env` file if present (via `godotenv`)
- Falls back to system environment variables
- Validates required fields on startup

---

## Running the Server

The main entrypoint is [`cmd/httpserver/main.go`](cmd/httpserver/main.go).

**Server Stack:**
- **Framework:** Echo (high performance HTTP router)
- **Middleware:** CORS, Gzip, Request ID, Recover, Security headers, Sentry
- **Error Handling:** Custom error handler maps domain errors to HTTP status codes
  - `EINVALID` → 400 Bad Request
  - `ENOTFOUND` → 404 Not Found
  - `EINTERNAL` → 500 Internal Server Error

**Server Initialization:**
```go
// 1. Load configuration
cfg, _ := config.LoadConfig()

// 2. Initialize infrastructure (database)
db, _ := postgres.NewConnection(postgres.Options{...})
userRepo := postgres.NewUserRepository(db)

// 3. Create use cases
userService := user.NewUsecase(userRepo, hashing.NewBcryptHasher())
authService := auth.NewUsecase(userRepo, postgres.NewLoginAttemptRepository(db), ...)

// 4. Inject into server
server := httpserver.Default()
server.UserService = userService
server.AuthService = authService
server.Addr = fmt.Sprintf(":%d", cfg.Port)
server.Start()
```

Default port is `8088` (configurable via `PORT` environment variable).

---

## Database Migrations

Migration files are located in [`migrations/`](migrations/) and managed using `sql-migrate`.

### Configuration

Database migration settings are in [`dbconfig.yml`](dbconfig.yml):
- Environment: `development`
- Migration directory: `migrations/`
- Database connection configured via environment variables

### Commands

**Run migrations:**
```shell
make db/migrate
# or directly:
go run ./cmd/migrate
```

**Create a new migration:**
```shell
sql-migrate new -env="development" create-your-migration-name
```

This creates a new file in `migrations/` with timestamp prefix (e.g., `20260202161352-create-users-table.sql`).

---

## Testing

The project includes comprehensive unit and integration tests using `testify` for assertions/mocking and `testcontainers` for database integration tests.

### Test Strategy

**1. Unit Tests** - Mock dependencies using `testify/mock`:
- [`user/usecase_test.go`](user/usecase_test.go) - Mocks repository/hasher to test user business logic
- [`httpserver/user_test.go`](httpserver/user_test.go) - Mocks user service to test user HTTP handlers

**2. Integration Tests** - Uses real PostgreSQL via `testcontainers`:
- [`httpserver/server_integration_test.go`](httpserver/server_integration_test.go) - Full stack testing setup
- [`postgres/postgres_test.go`](postgres/postgres_test.go) - Database layer testing utilities

### Running Tests

```shell
# Run all tests with coverage
make test

# Run specific test
go test ./user/... -v

# Run with coverage report
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Database Setup

Integration tests automatically:
1. Start PostgreSQL container via testcontainers
2. Run migrations on test database
3. Clean up after tests complete

See [`postgres/postgres_test.go`](postgres/postgres_test.go) for shared test utilities.

---

## Linting

Code quality is enforced using `golangci-lint` with configuration in [`.golangci.yml`](.golangci.yml).

```shell
make lint
```

This runs multiple linters including:
- `govet` - Static analysis
- `errcheck` - Unchecked errors
- And more...

Formatting is run separately:

```shell
make format
```

Generate Swagger docs:

```shell
make swag
# equivalent to:
swag init -g cmd/httpserver/main.go
```

---

## Docker

**Local PostgreSQL:** 
- Docker Compose configuration: [`tools/compose/docker-compose.yml`](tools/compose/docker-compose.yml)
- Starts PostgreSQL 15 on port `33062`
- Credentials configured via `.env` file

```shell
make local-db  # Start PostgreSQL
```

**Application Dockerfiles:**
- [`cmd/httpserver/Dockerfile`](cmd/httpserver/Dockerfile) - Main server
- [`cmd/migrate/Dockerfile`](cmd/migrate/Dockerfile) - Database migrations

---

## Project Structure

```
cmd/
  httpserver/          # HTTP server entrypoint
  migrate/            # Database migration entrypoint
auth/
  usecase.go          # Authentication use cases (login/refresh/oauth)
httpserver/
  server.go           # Echo server setup & middleware
  user.go             # User HTTP handlers
  auth.go             # Auth HTTP handlers
  health.go           # Health check handler
  request.go          # Request DTOs
  response.go         # Response DTOs
  *_test.go           # Unit tests
  *_integration_test.go  # Integration tests
postgres/
  postgres.go         # Database connection
  user.go             # User repository (GORM)
  login_attempt.go    # Login attempt repository (GORM)
  postgres_test.go    # Shared test utilities
user/
  user.go             # User entity + validation rules
  usecase.go          # User service/repository interfaces + implementation
  usecase_test.go     # Unit tests with mocks
errs/
  error.go            # Custom error types & codes
pkg/
  config/             # Configuration loader (envconfig)
  hashing/            # Password hashing
  jwt/                # JWT token provider
  oauth/google/       # Google OAuth provider
  sentry/             # Sentry error reporting
migrations/           # SQL migration files (sql-migrate)
tools/compose/        # Docker Compose files
```

### Key Design Patterns

**Dependency Injection:**
```go
// Domain defines interface
type Service interface { AddUser(context.Context, User) error }

// Infrastructure implements it
type Usecase struct { r Repository }

// Server depends on interface, not implementation
type Server struct { UserService user.Service }
```

**Error Handling:**
```go
// Domain errors with codes
var ErrInvalidName = errs.Errorf(errs.EINVALID, "invalid name")

// Custom HTTP error handler maps codes
// EINVALID -> 400, ENOTFOUND -> 404, EINTERNAL -> 500
```

**Testing:**
- Unit tests mock dependencies (Repository, Service)
- Integration tests use testcontainers for real database
- Shared test utilities for database setup/cleanup

---

## Available Make Commands

```shell
make run         # Start server with hot reload (Air)
make test        # Run all tests with coverage
make local-db    # Start PostgreSQL in Docker
make db/migrate  # Run database migrations
make lint        # Run golangci-lint
make format      # Format Go code (gofmt -w .)
make swag        # Generate Swagger docs
```
