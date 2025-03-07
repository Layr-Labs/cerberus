FROM golang:1.21 AS build

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

ARG APP_VERSION
RUN go build -ldflags "-X main.version=$APP_VERSION" -v -o bin/cerberus cmd/cerberus/main.go

FROM debian:latest
COPY bin/cerberus /cerberus

ENTRYPOINT [ "/cerberus"]
