FROM golang:1.22.1 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./common ./common/
COPY ./config/product-config.go ./config/
COPY ./domains/product ./domains/product/
COPY ./internal/http/chi/product ./internal/http/chi/product/
COPY ./cmd/product-service/main.go ./cmd/

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/main ./cmd/main.go

FROM debian:stable-slim

COPY ./db ./db/

COPY --from=build /bin/main /bin/main

# Install the postgresql-client package, which includes pg_isready
RUN apt-get update && \
    apt-get install -y postgresql-client && \
    rm -rf /var/lib/apt/lists/*

CMD ["./bin/main"]
