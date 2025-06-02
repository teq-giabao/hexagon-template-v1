# Hexagon Template

A Go project template following hexagonal architecture, featuring HTTP APIs, PostgreSQL and in-memory adapters, Sentry integration, and robust testing.

---

## Table of Contents

- [Architecture](#architecture)
- [Development](#development)
- [Configuration](#configuration)
- [Running the Server](#running-the-server)
- [Database Migrations](#database-migrations)
- [Testing](#testing)
- [Linting](#linting)
- [API Endpoints](#api-endpoints)
- [Docker](#docker)
- [Project Structure](#project-structure)

---

## Architecture

- **Domain Layer:** Business logic and interfaces (see [`domain/book`](domain/book/book.go)).
- **Adapters:** HTTP server ([`adapters/httpserver`](adapters/httpserver/)), PostgreSQL ([`adapters/postgrestore`](adapters/postgrestore/)), and in-memory ([`adapters/inmemstore`](adapters/inmemstore/)).
- **Configuration & Logging:** Centralized in [`pkg/config`](pkg/config/config.go) and [`pkg/logging`](pkg/logging/logger.go).
- **Error Reporting:** Sentry integration via [`pkg/sentry`](pkg/sentry/sentry.go).

---

## Development

### Init Local Environment

1. Copy `.env.example` to `.env` and update variables as needed.
2. Start local services:
    ```shell
    make local-db
    ```
3. Run the server:
    ```shell
    make run
    ```
4. Run unit tests:
    ```shell
    make test
    ```

---

## Configuration

- Environment variables are loaded from `.env` (see [`pkg/config/config.go`](pkg/config/config.go)).
- Example `.env`:
    ```
    APP_ENV=local
    PORT=8088
    DB_HOST=localhost
    DB_USER=root
    DB_PASS=123456
    DB_PORT=33062
    DB_NAME=teqlocal
    ```

---

## Running the Server

- The main entrypoint is [`cmd/httpserver/main.go`](cmd/httpserver/main.go).
- The server uses Echo and supports CORS, gzip, request ID, and Sentry middleware.
- Default port is `8088` (configurable).

---

## Database Migrations

- Migration files are in [`migrations/`](migrations/).
- To create a new migration:
    ```shell
    sql-migrate new -env="development" create-books-table
    ```
- To run migrations:
    ```shell
    make db/migrate
    ```

---

## Testing

- Unit and integration tests are in each adapter and domain package.
- Run all tests:
    ```shell
    make test
    ```
- Example: [`adapters/httpserver/server_test.go`](adapters/httpserver/server_test.go), [`adapters/postgrestore/book_store_test.go`](adapters/postgrestore/book_store_test.go)

---

## Linting

- Uses `golangci-lint` (see `.golangci.yml`).
    ```shell
    make lint
    ```

---

## API Endpoints

- **Health Check:** `GET /healthz`
- **Books:**
    - `POST /api/books` — Create a book (`{"isbn": "...", "name": "..."}`)
    - `GET /api/books/:id` — Get book by ISBN

---

## Docker

- Dockerfiles for server and migration in [`cmd/httpserver/Dockerfile`](cmd/httpserver/Dockerfile) and [`cmd/migrate/Dockerfile`](cmd/migrate/Dockerfile).
- Local PostgreSQL via Docker Compose: [`tools/compose/docker-compose.yml`](tools/compose/docker-compose.yml).

---

## Project Structure

```
adapters/
  httpserver/      # HTTP API server (Echo)
  postgrestore/    # PostgreSQL adapter
  inmemstore/      # In-memory adapter (BadgerDB)
  testutil/        # Test utilities
domain/
  book/            # Book domain logic and interfaces
pkg/
  config/          # Configuration loader
  logging/         # Logger setup
  sentry/          # Sentry integration
cmd/
  httpserver/      # Server entrypoint
  migrate/         # Migration entrypoint
migrations/        # SQL migration files
tools/compose/     # Docker Compose files
```
