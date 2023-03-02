FROM golang:1.19.1-buster
WORKDIR /app

RUN go install gotest.tools/gotestsum@v1.9.0

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .
RUN go build -o /merge-fixtures ./tooling/cmd/merge_fixtures.go

COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
