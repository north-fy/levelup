BINARY := bin/server
MIGRATE_BIN := bin/migrate
LINT_BIN := bin/golangci-lint
GO ?= go
export GOMODCACHE := $(CURDIR)/.gomodcache

.PHONY: run build test lint swagger migrate-up migrate-down migrate-up-clickhouse compose-up compose-down

run:
	$(GO) run ./cmd/server

build:
	$(GO) build -o $(BINARY) ./cmd/server

test:
	$(GO) test ./... -cover

lint: $(LINT_BIN)
	$(LINT_BIN) run ./...

$(LINT_BIN):
	@echo "golangci-lint not found in $(LINT_BIN). Install it:" 
	@echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b bin v2.12.2"

swagger:
	swag init -g cmd/server/main.go -o docs

migrate-up: $(MIGRATE_BIN)
	$(MIGRATE_BIN) -path migrations/postgres -database "$$(go run ./cmd/migrate -dsn)" up

migrate-down: $(MIGRATE_BIN)
	$(MIGRATE_BIN) -path migrations/postgres -database "$$(go run ./cmd/migrate -dsn)" down 1

migrate-up-clickhouse: $(MIGRATE_BIN)
	$(MIGRATE_BIN) -path migrations/clickhouse -database "$$(go run ./cmd/migrate -ch-dsn)" up

$(MIGRATE_BIN):
	$(GO) build -o $(MIGRATE_BIN) ./cmd/migrate

compose-up:
	docker-compose -f deploy/docker-compose.dev.yml up -d

compose-down:
	docker-compose -f deploy/docker-compose.dev.yml down