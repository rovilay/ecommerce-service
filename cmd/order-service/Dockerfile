FROM golang:1.22 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY ./common ./common/
COPY ./config/order-config.go ./config/
COPY ./domains/auth ./domains/auth/
COPY ./domains/order ./domains/order/
COPY ./internal/http/chi/order ./internal/http/chi/order/
COPY ./cmd/order-service/main.go ./cmd/

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/main ./cmd/main.go

FROM debian:stable-slim

COPY ./db ./db/

COPY --from=build /bin/main /bin/main

CMD ["./bin/main"]