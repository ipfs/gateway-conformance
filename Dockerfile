FROM golang:1.20-alpine
WORKDIR /app
ENV GATEWAY_CONFORMANCE_HOME=/app

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
RUN go build -ldflags="-X github.com/ipfs/gateway-conformance/tooling.Version=${VERSION}" -o ./gateway-conformance ./cmd/gateway-conformance

ENTRYPOINT ["/app/gateway-conformance"]
