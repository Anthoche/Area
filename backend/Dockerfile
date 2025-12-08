FROM golang:latest as build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /myapp ./src

FROM debian:bookworm-slim as run

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/* && \
    update-ca-certificates

COPY --from=build /myapp /myapp

WORKDIR /app

CMD ["/myapp"]
