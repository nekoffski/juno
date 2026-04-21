OAPI_CODEGEN  := $(shell go env GOPATH)/bin/oapi-codegen
REST_OPENAPI_SPEC := api/rest-openapi.yaml
REST_OAPI_CFG     := api/rest-oapi-codegen.yaml
VENV_DIR        := tests/.venv
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
	rm -rf bin/ coverage.out coverage *ut-coverage.txt $(VENV_DIR)
	find . -name '*.gen.go' -delete

unit-test:
	go test ./...

unit-test-verbose:
	go test -v ./...

unit-test-cover:
	go test -coverprofile=ut-coverage.txt ./...
	grep -v 'gen\.go:' ut-coverage.txt > ut-coverage.txt.tmp && mv ut-coverage.txt.tmp ut-coverage.txt

test-venv:
	python3 -m venv $(VENV_DIR)
	$(VENV_DIR)/bin/pip install -r tests/requirements.txt

functional-tests:
	./cicd/run-functional-tests.sh

docker-smoke-tests:
	./cicd/run-docker-smoke-tests.sh

generate-coverage:
	./cicd/generate-coverage.sh

.PHONY: all build generate gen-rest-api build lint fmt vet tidy clean test-venv unit-test unit-test-verbose unit-test-cover\
	functional-tests docker-smoke-tests generate-coverage



