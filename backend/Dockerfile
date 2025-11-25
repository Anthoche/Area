FROM golang:latest as build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /myapp ./src

FROM debian:bookworm-slim as run

COPY --from=build /myapp /myapp

WORKDIR /app

CMD ["/myapp"]
