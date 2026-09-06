# foreman -- autodev harness

MIGRATE_VERSION   ?= v4.19.0
GOIMPORTS_VERSION ?= latest
MIGRATE           := go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)

.PHONY: build run-api run-cli test test-e2e lint fmt \
        docker-build docker-up docker-down migrate-up migrate-down

build:
	go build -o bin/api ./cmd/api
	go build -o bin/cli ./cmd/cli

run-api:
	go run ./cmd/api

run-cli:
	go run ./cmd/cli $(ARGS)

test:
	go test -race ./...

test-e2e:
	go test -tags=e2e ./testes/...

lint:
	golangci-lint run

fmt:
	gofmt -w .
	go run golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION) -w .

docker-build:
	docker build -f deploy/Dockerfile -t foreman:latest .

docker-up:
	docker compose -f deploy/docker-compose.yml up -d --build

docker-down:
	docker compose -f deploy/docker-compose.yml down

migrate-up:
	@test -n "$(DATABASE_URL)" || { echo "DATABASE_URL is not set"; exit 1; }
	$(MIGRATE) -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	@test -n "$(DATABASE_URL)" || { echo "DATABASE_URL is not set"; exit 1; }
	$(MIGRATE) -path ./migrations -database "$(DATABASE_URL)" down 1
