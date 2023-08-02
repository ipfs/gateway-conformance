GIT_COMMIT := $(shell git rev-parse --short HEAD)
DIRTY_SUFFIX := $(shell test -n "`git status --porcelain`" && echo "-dirty" || true)
CLI_VERSION := dev-$(GIT_COMMIT)$(DIRTY_SUFFIX)

all: gateway-conformance

clean: clean-docker
	rm -f ./gateway-conformance
	rm -f *.ipns-record
	rm -f fixtures.car
	rm -f dnslinks.json
	rm -f ./reports/*

test-cargateway: provision-cargateway fixtures.car gateway-conformance
	./gateway-conformance test --json reports/output.json --gateway-url http://127.0.0.1:8040 --specs -subdomain-gateway

test-kubo-subdomains: provision-kubo gateway-conformance
	./kubo-config.example.sh
	./gateway-conformance test --json reports/output.json --gateway-url http://127.0.0.1:8080 --subdomain-url http://example.com:8080

test-kubo: provision-kubo gateway-conformance
	./gateway-conformance test --json reports/output.json --gateway-url http://127.0.0.1:8080 --specs -subdomain-gateway

provision-cargateway: ./fixtures.car
	# cd go-libipfs/examples/car && go install
	car -c ./fixtures.car &

provision-kubo:
	find ./fixtures -name '*.car' -exec ipfs dag import --stats --pin-roots=false {} \;
	find ./fixtures -name '*.ipns-record' -exec sh -c 'ipfs routing put --allow-offline /ipns/$$(basename -s .ipns-record "{}") "{}"' \;

# tools
fixtures.car: gateway-conformance
	./gateway-conformance extract-fixtures --merged=true --dir=.

gateway-conformance:
	go build -ldflags="-X github.com/ipfs/gateway-conformance/tooling.Version=$(CLI_VERSION)" -o ./gateway-conformance ./cmd/gateway-conformance

test-docker: docker fixtures.car gateway-conformance
	./gc test

./reports/output.xml: ./reports/output.json
	jq -ns 'inputs' ./reports/output.json > ./reports/output.json.alt
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" ghcr.io/pl-strflt/saxon:v1 -json:"./reports/output.json.alt" -xsl:/etc/gotest.xsl -o:"./reports/output.xml"

./reports/output.html: ./reports/output.xml
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" ghcr.io/pl-strflt/saxon:v1 -s:./reports/output.xml -xsl:/etc/junit-noframes-saxon.xsl -o:./reports/output.html
	open ./reports/output.html

docker:
	docker build --build-arg VERSION="$(CLI_VERSION)" -t gateway-conformance .

clean-docker:
	@if command -v docker >/dev/null 2>&1 && docker image inspect gateway-conformance >/dev/null 2>&1; then \
        docker image rm gateway-conformance; \
    fi

.PHONY: gateway-conformance
