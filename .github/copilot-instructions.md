# Hexagon Go Project - AI Coding Guidelines

## Architecture Overview

This is a Go REST API following **hexagonal/clean architecture** with clear separation of concerns:

**3-Layer Architecture:**
- **Domain layer** (`contact/`): Business entities, validation logic, and interface definitions (`Service`, `Repository`)
- **Application layer** (`contact/usecase.go`): Business logic implementation (use cases) that coordinates between domain and infrastructure
- **Infrastructure layer**: Adapters for external dependencies
  - `httpserver/` - Echo web framework HTTP handlers
  - `postgres/` - GORM database implementation
  - `pkg/` - Cross-cutting concerns (config, sentry, logging)

**Dependency injection pattern:**
```go
// Domain defines interfaces
type Service interface { AddContact(Contact) error }
type Repository interface { CreateContact(Contact) error }

// Usecase implements Service, depends on Repository
type Usecase struct { r Repository }

// Infrastructure implements Repository
type ContactRepository struct { db *gorm.DB }

// Server depends on Service (not implementation)
type Server struct { ContactService contact.Service }
```

## Development Workflow

**Essential commands:**
```bash
make local-db    # Start PostgreSQL in Docker (port 33062)
make run         # Hot reload with Air (.air.toml)
make test        # Run all tests with coverage
make lint        # golangci-lint validation
make db/migrate  # Apply sql-migrate migrations
```

**Hot reload setup (Air):**
- Watches `*.go` files (excludes `*_test.go`)
- Builds to `tmp/main` from `cmd/httpserver/`
- Auto-restarts on file changes

**Testing strategy:**
1. **Unit tests** - Mock dependencies using `testify/mock`
   - `contact/usecase_test.go` - Mocks Repository
   - `httpserver/contact_test.go` - Mocks Service
2. **Integration tests** - Uses `testcontainers` for real PostgreSQL
   - `httpserver/contact_integration_test.go` - Full stack test
   - `postgres/contact_test.go` - Database layer tests

## Code Patterns & Conventions

**Server initialization (see `cmd/httpserver/main.go`):**
```go
// 1. Load config from environment/.env
cfg, _ := config.LoadConfig()

// 2. Initialize infrastructure
db, _ := postgres.NewConnection(postgres.Options{...})
contactRepo := postgres.NewContactRepository(db)

// 3. Create use cases
contactService := contact.NewUsecase(contactRepo)

// 4. Inject into server
server := httpserver.Default()
server.ContactService = contactService
server.Addr = fmt.Sprintf(":%d", cfg.Port)
server.Start()
```

**Route registration pattern (`httpserver/`):**
- Routes registered in `Default()` constructor
- Dedicated methods per domain: `RegisterContactRoutes()`, `RegisterHealthRoutes()`
- Middleware applied globally in `RegisterGlobalMiddlewares()`

**Error handling (`errs/error.go`):**
```go
// Define domain errors with codes
var ErrInvalidName = errs.Errorf(errs.EINVALID, "invalid name")

// Custom handler maps codes to HTTP status
// EINVALID -> 400, ENOTFOUND -> 404, EINTERNAL -> 500
func customHTTPErrorHandler(err error, c echo.Context) {...}
```

**Database testing (`postgres/*_test.go`):**
```go
// Shared test setup pattern
db := CreateConnection(t, dbName, dbUser, dbPass)
MigrateTestDatabase(t, db, "../migrations")
cleanupContactDatabase(t, db) // Clean state per test
```

**Integration test pattern (`httpserver/*_integration_test.go`):**
```go
db := MustCreateTestDatabase(t)  // testcontainers PostgreSQL
MigrateTestDatabase(t, db, "../migrations")
server := MustCreateServer(t, db)  // Wire real dependencies
server.Router.ServeHTTP(rec, req)  // Test via HTTP
```

## Key Files & Locations

**Domain:**
- `contact/contact.go` - Entity with validation
- `contact/usecase.go` - Service & Repository interfaces + implementation

**Infrastructure:**
- `httpserver/server.go` - Echo setup, middleware, error handling
- `httpserver/contact.go` - HTTP handlers for contact routes
- `postgres/contact.go` - GORM models & Repository implementation
- `postgres/postgres_test.go` - Shared test helpers (testcontainers setup)

**Configuration:**
- `.env` - Local environment variables (gitignored)
- `pkg/config/config.go` - Loads config via envconfig + godotenv
- `dbconfig.yml` - sql-migrate database configuration
- `.air.toml` - Hot reload configuration

**Testing:**
- `*_test.go` - Unit tests with mocks (same package as `_test` suffix)
- `*_integration_test.go` - Integration tests with testcontainers
