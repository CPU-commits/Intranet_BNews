FROM golang:latest

RUN export GO111MODULE="on"

RUN go install github.com/githubnemo/CompileDaemon@latest

WORKDIR /app

COPY . /app

RUN go mod download

ENTRYPOINT CompileDaemon -polling --build="go build src/main.go" --command=./main