FROM golang:1.19.1-buster
WORKDIR /app
ENV TEST_PATH=/app

RUN go install gotest.tools/gotestsum@v1.9.0

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .
RUN go build -o /entrypoint ./entrypoint.go

ENTRYPOINT ["/entrypoint"]
