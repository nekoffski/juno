OAPI_CODEGEN  := $(shell go env GOPATH)/bin/oapi-codegen
REST_OPENAPI_SPEC := api/rest-openapi.yaml
REST_OAPI_CFG     := api/rest-oapi-codegen.yaml
VENV_DIR        := tests/.venv
ENV_FILE        := $(or $(ENV_FILE),conf/.env.example)
CONDUCTOR_CFG   := $(or $(CONDUCTOR_CFG),conf/conductor.yaml)
TARGETS = juno-server juno-web juno-mcp juno-conductor juno-lan-agent
GO = go
GOLANGCI_LINT := $(shell $(GO) env GOPATH)/bin/golangci-lint


all: build

generate: gen-rest-api

gen-rest-api:
	$(OAPI_CODEGEN) --config $(REST_OAPI_CFG) $(REST_OPENAPI_SPEC)

build: generate
	for target in $(TARGETS); do \
		$(GO) build -o bin/$$target ./cmd/$$target; \
	done

coverage-build:
	for target in $(TARGETS); do \
		$(GO) build -cover -o bin/$$target ./cmd/$$target; \
	done

lint:
	$(GOLANGCI_LINT) run ./...

fmt:
	$(GO) fmt ./...

test-fmt:
	test -z "$(shell gofmt -l .)"

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -rf bin/ coverage.out coverage *coverage.txt $(VENV_DIR)
	find . -name '*.gen.go' -delete

unit-test:
	go test ./...

unit-test-verbose:
	go test -v ./...

unit-test-cover:
	go test -coverprofile=coverage.txt ./...
	grep -v 'gen\.go:' coverage.txt > coverage.txt.tmp && mv coverage.txt.tmp coverage.txt

.PHONY: all build generate gen-rest-api build lint fmt vet tidy clean


# run-server:
# 	env $(shell grep -v '^#' $(ENV_FILE) | grep '=' | xargs) go run $(CMD_SERVER)

# run-web:
# 	env $(shell grep -v '^#' $(ENV_FILE) | grep '=' | xargs) go run $(CMD_WEB)

# run-mcp:
# 	env $(shell grep -v '^#' $(ENV_FILE) | grep '=' | xargs) go run $(CMD_MCP)

# run-conductor:
# 	env $(shell grep -v '^#' $(ENV_FILE) | grep '=' | xargs) go run $(CMD_CONDUCTOR) -config $(CONDUCTOR_CFG)

# run-lan-agent:
# 	env $(shell grep -v '^#' $(ENV_FILE) | grep '=' | xargs) go run $(CMD_LAN)

# docker-build-lan-agent:
# 	docker build \
# 		$(shell grep -v '^#' $(ENV_FILE) | grep '=' | sed 's/^/--build-arg /') \
# 		-f Dockerfile.lan -t juno-lan-agent .

# docker-run-lan-agent:
# 	docker run -d --name juno-lan-agent --network host --env-file $(ENV_FILE) juno-lan-agent

# docker-stop-lan-agent:
# 	docker stop juno-lan-agent && docker rm juno-lan-agent



# unit-coverage: unit-test-cover-ci
# 	bash cicd/generate-coverage-reports.sh coverage.txt coverage/unit

# integration-test-setup:
# 	bash cicd/integration-test-setup.sh

# integration-test-run:
# 	bash cicd/integration-test-run.sh

# integration-test-teardown:
# 	bash cicd/integration-test-teardown.sh

# test-env-up:
# 	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) up -d postgres

# test-env-down:
# 	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) down

# docker-build:
# 	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) build

# docker-up:
# 	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) up -d

# docker-up-interactive:
# 	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) up

# docker-down:
# 	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) down

# docker-logs:
# 	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) logs -f

# docker-restart: docker-down docker-up

# docker-up-%:
# 	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) up -d $*

# docker-down-%:
# 	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) stop $*

# docker-logs-%:
# 	ENV_FILE=$(ENV_FILE) docker compose --env-file $(ENV_FILE) logs -f $*

# integration-test:
# 	env $(shell grep -v '^#' $(ENV_FILE) | xargs) $(VENV_DIR)/bin/pytest tests/ -v

# test-venv:
# 	python3 -m venv $(VENV_DIR)
# 	$(VENV_DIR)/bin/pip install -r tests/requirements.txt

