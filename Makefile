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

# tools
fixtures.car: gateway-conformance
	./gateway-conformance extract-fixtures --merged=true --dir=.

gateway-conformance:
	go build -o ./gateway-conformance ./cmd/gateway-conformance

test-docker: docker fixtures.car gateway-conformance
	./gc test

output.xml: test-kubo
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" --entrypoint "/bin/bash" ghcr.io/pl-strflt/saxon:v1 -c """
		java -jar /opt/SaxonHE11-5J/saxon-he-11.5.jar -s:<(jq -s '.' output.json) -xsl:/etc/gotest.xsl -o:output.xml
	"""

output.html: output.xml
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" ghcr.io/pl-strflt/saxon:v1 -s:output.xml -xsl:/etc/junit-noframes-saxon.xsl -o:output.html
	open ./output.html

docker:
	docker build -t gateway-conformance .

.PHONY: gateway-conformance
