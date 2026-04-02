APP          := juno
CMD          := ./cmd/juno
BIN_DIR      := bin
OAPI_CODEGEN := $(shell go env GOPATH)/bin/oapi-codegen
REST_OPENAPI_SPEC := api/rest-openapi.yaml
REST_OAPI_CFG     := api/rest-oapi-codegen.yaml
VENV_DIR     := tests/.venv
ENV_FILE     := $(or $(ENV_FILE),.env.example)
.PHONY: all build run unit-test unit-test-verbose unit-test-cover unit-test-cover-ci unit-coverage \
        integration-test-coverage clean lint fmt vet tidy generate gen-api \
        docker-build docker-up docker-up-interactive docker-down docker-logs docker-restart \
        integration-test venv test-env-up test-env-down

all: build

generate: gen-rest-api

gen-rest-api:
	$(OAPI_CODEGEN) --config $(REST_OAPI_CFG) $(REST_OPENAPI_SPEC)

build: generate
	go build -o $(BIN_DIR)/$(APP) $(CMD)

run:
	env $(shell grep -v '^#' $(ENV_FILE) | grep '=' | xargs) go run $(CMD)

unit-test:
	go test ./...

unit-test-verbose:
	go test -v ./...

unit-test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

unit-test-cover-ci:
	go test -coverprofile=unit-coverage.out ./...

unit-coverage: unit-test-cover-ci
	bash cicd/generate-coverage-reports.sh unit-coverage.out coverage/unit

integration-test-coverage:
	bash cicd/integration-test-with-coverage.sh

test-env-up:
	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) up -d postgres

test-env-down:
	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) down

docker-build:
	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) build

docker-up:
	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) up -d

docker-up-interactive:
	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) up

docker-down:
	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) down

docker-logs:
	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) logs -f

docker-restart: docker-down docker-up

docker-up-%:
	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) up -d $*

docker-down-%:
	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) stop $*

docker-logs-%:
	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) logs -f $*

integration-test:
	env $(shell grep -v '^#' $(ENV_FILE) | xargs) $(VENV_DIR)/bin/pytest tests/ -v

test-venv:
	python3 -m venv $(VENV_DIR)
	$(VENV_DIR)/bin/pip install -r tests/requirements.txt

clean:
	rm -rf $(BIN_DIR)/ coverage.out $(VENV_DIR)
	find . -name '*.gen.go' -delete

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

tidy:
	go mod tidy
