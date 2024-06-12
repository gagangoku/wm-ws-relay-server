# Run supergroupmod in docker:
#
# Build as docker build -t relay-server .
# Run as docker run -it --rm -p 8080:8080 -e EXTERNAL_ENDPOINT=ws://localhost:8080 -e LISTEN_PORT=8080 relay-server
#

FROM golang:1.22-alpine

WORKDIR /app

RUN apk add bash

COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go mod tidy

COPY . .

WORKDIR /app
RUN go mod tidy

RUN go build -o relay-server.out

WORKDIR /app

ENTRYPOINT [ "./docker-entrypoint.sh" ]
# CMD [ "sleep", "infinity" ]

# Ignore below, its only for quick debugging
# WORKDIR /app
# COPY . .
# ENTRYPOINT [ "bash", "./docker-entrypoint.sh" ]
