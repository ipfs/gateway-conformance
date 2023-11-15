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
	go build \
		-ldflags="-X github.com/ipfs/gateway-conformance/tooling.Version=$(CLI_VERSION)" \
		-o ./gateway-conformance \
		./cmd/gateway-conformance

test-docker: docker fixtures.car gateway-conformance
	./gc test

./reports/output.xml: ./reports/output.json
	jq -ns 'inputs' ./reports/output.json > ./reports/output.json.alt
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" ghcr.io/pl-strflt/saxon:v1 -json:"./reports/output.json.alt" -xsl:/etc/gotest.xsl -o:"./reports/output.xml"

./reports/output.html: ./reports/output.xml
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" ghcr.io/pl-strflt/saxon:v1 -s:./reports/output.xml -xsl:/etc/junit-noframes-saxon.xsl -o:./reports/output.html
	open ./reports/output.html

./reports/output.md: ./reports/output.xml
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" ghcr.io/pl-strflt/saxon:v1 -s:./reports/output.xml -xsl:/etc/junit-summary.xsl -o:./reports/output.md

docker:
	docker build --build-arg VERSION="$(CLI_VERSION)" -t gateway-conformance .

clean-docker:
	@if command -v docker >/dev/null 2>&1 && docker image inspect gateway-conformance >/dev/null 2>&1; then \
        docker image rm gateway-conformance; \
    fi

# dashboard
raw_artifacts:
	cat REPOSITORIES | xargs ./munge_download.sh ./artifacts

artifacts: raw_artifacts
	find ./artifacts -name '*.json' -exec sh -c 'cat "{}" | node ./munge.js > out && mv out "{}"' \;

aggregates.db: artifacts
	rm -f ./aggregates.db
	node ./munge_sql.js ./aggregates.db ./artifacts/*.json

www_content: aggregates.db
	node ./munge_aggregates.js ./aggregates.db ./www

www: www_content
	cd www && hugo --minify $(if ${OUTPUT_BASE_URL},--baseURL ${OUTPUT_BASE_URL})

.PHONY: gateway-conformance
