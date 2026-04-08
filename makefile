APP_SERVER    := juno-server
APP_WEB       := juno-web
APP_MCP       := juno-mcp
APP_CONDUCTOR := juno-conductor
APP_LAN       := juno-lan-agent
CMD_SERVER    := ./cmd/juno-server
CMD_WEB       := ./cmd/juno-web
CMD_MCP       := ./cmd/juno-mcp
CMD_CONDUCTOR := ./cmd/juno-conductor
CMD_LAN       := ./cmd/juno-lan-agent
BIN_DIR       := bin
OAPI_CODEGEN  := $(shell go env GOPATH)/bin/oapi-codegen
REST_OPENAPI_SPEC := api/rest-openapi.yaml
REST_OAPI_CFG     := api/rest-oapi-codegen.yaml
VENV_DIR        := tests/.venv
ENV_FILE        := $(or $(ENV_FILE),conf/.env.example)
CONDUCTOR_CFG   := $(or $(CONDUCTOR_CFG),conf/conductor.yaml)

.PHONY: all build run run-web run-mcp run-conductor run-lan-agent unit-test unit-test-verbose unit-test-cover unit-test-cover-ci unit-coverage \
        integration-test-setup integration-test-run integration-test-teardown \
        clean lint fmt vet tidy generate gen-api \
        docker-build docker-up docker-up-interactive docker-down docker-logs docker-restart \
        integration-test venv test-env-up test-env-down

all: build

generate: gen-rest-api

gen-rest-api:
	$(OAPI_CODEGEN) --config $(REST_OAPI_CFG) $(REST_OPENAPI_SPEC)

build: generate
	go build -o $(BIN_DIR)/$(APP_SERVER) $(CMD_SERVER)
	go build -o $(BIN_DIR)/$(APP_WEB) $(CMD_WEB)
	go build -o $(BIN_DIR)/$(APP_MCP) $(CMD_MCP)
	go build -o $(BIN_DIR)/$(APP_CONDUCTOR) $(CMD_CONDUCTOR)
	go build -o $(BIN_DIR)/$(APP_LAN) $(CMD_LAN)

run-server:
	env $(shell grep -v '^#' $(ENV_FILE) | grep '=' | xargs) go run $(CMD_SERVER)

run-web:
	env $(shell grep -v '^#' $(ENV_FILE) | grep '=' | xargs) go run $(CMD_WEB)

run-mcp:
	env $(shell grep -v '^#' $(ENV_FILE) | grep '=' | xargs) go run $(CMD_MCP)

run-conductor:
	env $(shell grep -v '^#' $(ENV_FILE) | grep '=' | xargs) go run $(CMD_CONDUCTOR) -config $(CONDUCTOR_CFG)

run-lan-agent:
	env $(shell grep -v '^#' $(ENV_FILE) | grep '=' | xargs) go run $(CMD_LAN)

unit-test:
	go test ./...

unit-test-verbose:
	go test -v ./...

unit-test-cover:
	go test -coverprofile=coverage.out ./...
	grep -v 'gen\.go:' coverage.out > coverage.out.tmp && mv coverage.out.tmp coverage.out
	go tool cover -func=coverage.out

unit-test-cover-ci:
	go test -coverprofile=coverage.txt ./...
	grep -v 'gen\.go:' coverage.txt > coverage.txt.tmp && mv coverage.txt.tmp coverage.txt

unit-coverage: unit-test-cover-ci
	bash cicd/generate-coverage-reports.sh coverage.txt coverage/unit

integration-test-setup:
	bash cicd/integration-test-setup.sh

integration-test-run:
	bash cicd/integration-test-run.sh

integration-test-teardown:
	bash cicd/integration-test-teardown.sh

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
	rm -rf $(BIN_DIR)/ coverage.out coverage *coverage.txt $(VENV_DIR)
	find . -name '*.gen.go' -delete

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

tidy:
	go mod tidy
