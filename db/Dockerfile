FROM golang:1.22.1 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./config/product-config.go ./config/
COPY ./db/scripts/migration.go ./migration.go

# compile migration script
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/migration ./migration.go

FROM debian:stable-slim

COPY ./db/migrations ./db/migrations
COPY --from=build /bin/migration /bin/migration

# Install the postgresql-client package, which includes pg_isready
RUN apt-get update && \
    apt-get install -y postgresql-client && \
    rm -rf /var/lib/apt/lists/*

# run migrations
# CMD ["./bin/migration"]
