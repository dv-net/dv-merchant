.PHONY:
.SILENT:
.DEFAULT_GOAL := run

include .docker/dev/.env

MIGRATIONS_DIR = ./sql/postgres/migrations/
SEEDS_DIR = ./sql/postgres/seeds/

VERSION ?= $(strip $(shell ./scripts/version.sh))
VERSION_NUMBER := $(strip $(shell ./scripts/version.sh number))
COMMIT_HASH := $(shell git rev-parse --short HEAD)

OUT_BIN ?= ./.bin/dv-merchant
CUSTOM_CI_LINTER ?= ./.bin/dv-golangci-lint
GO_LDFLAGS ?=
GO_OPT_BASE := -ldflags "-X main.version=$(VERSION) $(GO_LDFLAGS) -X main.commitHash=$(COMMIT_HASH)"

BUILD_ENV := CGO_ENABLED=0
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S), Linux)
	BUILD_ENV += GOOS=linux
endif
ifeq ($(UNAME_S), Darwin)
	BUILD_ENV += GOOS=darwin
endif

UNAME_P := $(shell uname -p)
ifeq ($(UNAME_P),x86_64)
	BUILD_ENV += GOARCH=amd64
endif
ifneq ($(filter arm%,$(UNAME_P)),)
	BUILD_ENV += GOARCH=arm64
endif

## Build:

build:
	 go build $(GO_OPT_BASE) -o $(OUT_BIN) ./cmd/app

build_plugins:
	go tool golangci-lint custom -v

run: build
	$(OUT_BIN) $(filter-out $@,$(MAKECMDGOALS))

start:
	go run ./cmd/app start $(filter-out $@,$(MAKECMDGOALS))

test:
	go test ./...

version: ## Version of the project to built
	echo $(VERSION)

version-number: ## Version number of the project to be built
	echo $(VERSION_NUMBER)

## Lint:
lint: build_plugins
	$(CUSTOM_CI_LINTER) run --show-stats

fmt:
	go tool gofumpt -l -w .

genmocks:
	mockery

gensql:
	@go run generators/blockchain/main.go
	cd sql && pgxgen -pgxgen-config=pgxgen-postgres.yaml -sqlc-config=sqlc-postgres.yaml crud
	cd sql && pgxgen -pgxgen-config=pgxgen-postgres.yaml -sqlc-config=sqlc-postgres.yaml sqlc generate

migrate:
	go run ./cmd/app migrate $(filter-out $@,$(MAKECMDGOALS))

db-create-migration:
	migrate create -ext sql -dir "$(MIGRATIONS_DIR)" $(filter-out $@,$(MAKECMDGOALS))

db-create-seed:
	migrate create -ext sql -dir "$(SEEDS_DIR)" $(filter-out $@,$(MAKECMDGOALS))

db-seed:
	go run ./cmd/app seed --driver "$(APP_DB__ENGINE)" $(filter-out $@,$(MAKECMDGOALS))

swag-gen:
	swag fmt
	swag init --parseDependency --parseInternal -g ./cmd/app/main.go

swag-gen-external:
	swag fmt
	swag init --parseDependency --parseInternal --generalInfo ../../../../../cmd/app/main.go --dir ./internal/delivery/http/handlers/external --output ./docs/external

genenvs:
	go run ./cmd/app config genenvs

builds:
	$(BUILD_ENV) && go build $(GO_OPT_BASE) -o $(OUT_BIN) ./cmd/app

# Empty goals trap
%:
	@: