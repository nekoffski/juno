FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN /go/bin/oapi-codegen --config api/rest-oapi-codegen.yaml api/rest-openapi.yaml > internal/rest/api.gen.go

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/juno-server ./cmd/juno-server
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/juno-web ./cmd/juno-web
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/juno-mcp ./cmd/juno-mcp
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/juno-conductor ./cmd/juno-conductor

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/bin/juno-server .
COPY --from=builder /app/bin/juno-web .
COPY --from=builder /app/bin/juno-mcp .
COPY --from=builder /app/bin/juno-conductor .
COPY --from=builder /app/conf/conductor.yaml conf/conductor.yaml

CMD ["./juno-conductor", "-config", "conf/conductor.yaml"]
