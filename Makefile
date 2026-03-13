.PHONY: run test test/ci local-db lint format swag db/migrate db/seed docker/up docker/down docker/logs docker/migrate docker/seed

TEST_PKGS := $(shell go list ./... | grep -Ev '(cmd|testutil|mocks)')
DOCKER_COMPOSE_FULL := docker-compose --env-file ./.env -f ./tools/compose/docker-compose.full.yml

run:
	air -c .air.toml

test:
	go clean -testcache
	go test -cover ./...

test/ci:
	go clean -testcache
	go run gotest.tools/gotestsum@latest \
		--junitfile unit-tests.xml \
		--format pkgname -- \
		-coverprofile=coverage.out \
		$(TEST_PKGS)

local-db:
	docker-compose --env-file ./.env -f ./tools/compose/docker-compose.yml down
	docker-compose --env-file ./.env -f ./tools/compose/docker-compose.yml up -d

lint:
	golangci-lint version
	golangci-lint run

format:
	gofmt -w .

swag:
	swag init -g cmd/httpserver/main.go

db/migrate:
	go run ./cmd/migrate

db/seed:
	go run ./cmd/seed

docker/up:
	$(DOCKER_COMPOSE_FULL) up -d db
	$(DOCKER_COMPOSE_FULL) run --rm migrate
	$(DOCKER_COMPOSE_FULL) run --rm seed
	$(DOCKER_COMPOSE_FULL) up -d api

docker/down:
	$(DOCKER_COMPOSE_FULL) down

docker/logs:
	$(DOCKER_COMPOSE_FULL) logs -f api

docker/migrate:
	$(DOCKER_COMPOSE_FULL) run --rm migrate

docker/seed:
	$(DOCKER_COMPOSE_FULL) run --rm seed
