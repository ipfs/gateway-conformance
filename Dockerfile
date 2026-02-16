FROM golang:1.26-alpine
WORKDIR /app
ENV GATEWAY_CONFORMANCE_HOME=/app \
    GOCACHE=/go/cache

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
RUN go build -ldflags="-X github.com/ipfs/gateway-conformance/tooling.Version=${VERSION}" -o ./gateway-conformance ./cmd/gateway-conformance

# Relaxed perms for cache dir to allow running under regular user
RUN mkdir -p $GOCACHE && chmod -R 777 $GOCACHE

ENTRYPOINT ["/app/gateway-conformance"]
