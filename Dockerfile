FROM golang:1.19.3


ENV USERNAME="admin"
ENV PASSWORD="admin"

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build


CMD go run github.com/prisma/prisma-client-go migrate dev --name init; ./kontrakt-server