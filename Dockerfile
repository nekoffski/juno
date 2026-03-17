FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN /go/bin/oapi-codegen --config api/rest-oapi-codegen.yaml api/rest-openapi.yaml > internal/rest/api.gen.go

RUN CGO_ENABLED=0 GOOS=linux go build -o bin/juno ./cmd/juno

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/bin/juno .

CMD ["./juno"]
