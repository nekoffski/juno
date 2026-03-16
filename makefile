APP          := juno
CMD          := ./cmd/juno
BIN_DIR      := bin
OAPI_CODEGEN := $(shell go env GOPATH)/bin/oapi-codegen
REST_OPENAPI_SPEC := api/rest-openapi.yaml
REST_OAPI_CFG     := api/rest-oapi-codegen.yaml

.PHONY: all build run test clean lint fmt vet tidy generate gen-api

all: build

generate: gen-rest-api

gen-rest-api:
	$(OAPI_CODEGEN) --config $(REST_OAPI_CFG) $(REST_OPENAPI_SPEC)

build: generate
	go build -o $(BIN_DIR)/$(APP) $(CMD)

run:
	go run $(CMD)

test:
	go test ./...

test-verbose:
	go test -v ./...

clean:
	rm -rf $(BIN_DIR)/
	find . -name '*.gen.go' -delete

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

tidy:
	go mod tidy
