.PHONY: run test local-db lint db/migrate

run:
	air -c .air.toml

test:
	go clean -testcache
	go test -cover ./...

local-db:
	docker-compose --env-file ./.env -f ./tools/compose/docker-compose.yml down
	docker-compose --env-file ./.env -f ./tools/compose/docker-compose.yml up -d

lint:
	golangci-lint version
	golangci-lint run

db/migrate:
	go run ./cmd/migrate