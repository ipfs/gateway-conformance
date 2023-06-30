all: test-kubo

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
	find ./fixtures -name '*.car' -exec ipfs dag import {} \;
	find ./fixtures -name '*.ipns-record' -exec sh -c 'ipfs routing put --allow-offline /ipns/$$(basename -s .ipns-record "{}") "{}"' \;

# tools
fixtures.car: gateway-conformance
	./gateway-conformance extract-fixtures --merged=true --dir=.

gateway-conformance:
	go build -o ./gateway-conformance ./cmd/gateway-conformance

test-docker: docker fixtures.car gateway-conformance
	./gc test

./reports/output.xml: ./reports/output.json
	jq -ns 'inputs' ./reports/output.json > ./reports/output.json.alt
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" ghcr.io/pl-strflt/saxon:v1 -json:"./reports/output.json.alt" -xsl:/etc/gotest.xsl -o:"./reports/output.xml"
	
./reports/output.html: ./reports/output.xml
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" ghcr.io/pl-strflt/saxon:v1 -s:./reports/output.xml -xsl:/etc/junit-noframes-saxon.xsl -o:./reports/output.html
	open ./reports/output.html

docker:
	docker build -t gateway-conformance .

.PHONY: gateway-conformance
