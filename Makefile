GIT_COMMIT := $(shell git rev-parse --short HEAD)
DIRTY_SUFFIX := $(shell test -n "`git status --porcelain`" && echo "-dirty" || true)
CLI_VERSION := dev-$(GIT_COMMIT)$(DIRTY_SUFFIX)
KUBO_VERSION ?= latest
KUBO_DOCKER_NAME ?= kubo-$(KUBO_VERSION)-gateway-conformance

all: gateway-conformance

clean: clean-docker
	rm -f ./gateway-conformance
	rm -f *.ipns-record
	rm -f fixtures.car
	rm -f dnslinks.*
	rm -f dnslink*.yml
	rm -rf ./reports/*

test-cargateway: provision-cargateway fixtures.car gateway-conformance
	./gateway-conformance test --json reports/output.json --gateway-url http://127.0.0.1:8040 --specs -subdomain-gateway

test-kubo-subdomains: provision-kubo gateway-conformance
	./kubo-config.example.sh
	./gateway-conformance test --json reports/output.json --gateway-url http://127.0.0.1:8080 --subdomain-url http://example.com:8080

test-kubo: provision-kubo gateway-conformance
	./gateway-conformance test --json reports/output.json --gateway-url http://127.0.0.1:8080 --specs -subdomain-gateway

provision-cargateway: ./fixtures.car
	car -c ./fixtures.car &

provision-kubo:
	find ./fixtures -name '*.car' -exec ipfs dag import --stats --pin-roots=false {} \;
	find ./fixtures -name '*.ipns-record' -exec sh -c 'ipfs routing put --allow-offline /ipns/$$(basename -s .ipns-record "{}" | cut -d'_' -f1) "{}"' \;

#start-kubo-docker: stop-kubo-docker gateway-conformance
#	./gateway-conformance extract-fixtures --dir=.temp/fixtures
#	docker pull ipfs/kubo:$(KUBO_VERSION)
#	docker run -d --rm --net=host --name $(KUBO_DOCKER_NAME) -v "$(shell realpath .temp/fixtures)":/fixtures -v kubo-config.example.sh:/container-init.d/001-config.sh ipfs/kubo:$(KUBO_VERSION) daemon --init --offline
#	@until docker exec $(KUBO_DOCKER_NAME) ipfs --api=/ip4/127.0.0.1/tcp/5001 dag stat /ipfs/QmUNLLsPACCz1vLxQVkXqqLX5R1X345qqfHbsf67hvA3Nn >/dev/null 2>&1; do sleep 0.1; done
#	docker exec $(KUBO_DOCKER_NAME) find /fixtures -name '*.car' -exec ipfs dag import --stats --pin-roots=false {} \;
#	docker exec $(KUBO_DOCKER_NAME) find /fixtures -name '*.ipns-record' -exec sh -c 'ipfs routing put --allow-offline /ipns/$$(basename -s .ipns-record "{}" | cut -d'_' -f1) "{}"' \;
#	TODO: provision Kubo config at Gateway.PublicGateways  to have subdomain gateway on example.com and also enable inlining on localhost
#	      See: https://github.com/ipfs/kubo/blob/a07852a3f0294974b802923fb136885ad077384e/.github/workflows/gateway-conformance.yml#L22-L34
#	      (this is not as trivial as it sounds because Kubo does not apply config inrealtime, and a restart is required.)

stop-kubo-docker: clean
	docker stop $(KUBO_DOCKER_NAME) || true
	docker rm -f $(KUBO_DOCKER_NAME) || true

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
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" ghcr.io/ipdxco/saxon:v1 -json:"./reports/output.json.alt" -xsl:/etc/gotest.xsl -o:"./reports/output.xml"

./reports/output.html: ./reports/output.xml
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" ghcr.io/ipdxco/saxon:v1 -s:./reports/output.xml -xsl:/etc/junit-noframes-saxon.xsl -o:./reports/output.html
	open ./reports/output.html

./reports/output.md: ./reports/output.xml
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" ghcr.io/ipdxco/saxon:v1 -s:./reports/output.xml -xsl:/etc/junit-summary.xsl -o:./reports/output.md

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

website_content: aggregates.db
	node ./munge_aggregates.js ./aggregates.db ./www

website: website_content
	cd www && hugo --minify $(if ${OUTPUT_BASE_URL},--baseURL ${OUTPUT_BASE_URL})

.PHONY: gateway-conformance
