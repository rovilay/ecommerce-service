FROM golang:1.22.1 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./common ./common/
COPY ./config/inventory-config.go ./config/
COPY ./domains/inventory ./domains/inventory/
COPY ./internal/http/chi/inventory ./internal/http/chi/inventory/
COPY ./cmd/inventory-service/main.go ./cmd/

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/main ./cmd/main.go

FROM debian:stable-slim

COPY ./db ./db/

COPY --from=build /bin/main /bin/main

CMD ["./bin/main"]
