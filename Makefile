provision-cargateway: ./fixtures.car
	# cd go-libipfs/examples/car && go install
	car -c ./fixtures.car &

test-cargateway: provision-cargateway
	GATEWAY_URL=http://127.0.0.1:8040 make _test

provision-kubo:
	find ./fixtures -name '*.car' -exec ipfs dag import {} \;

test-kubo-subdomains: provision-kubo gateway-conformance
	ipfs config --json Gateway.PublicGateways '{	\
		"example.com": {							\
			"UseSubdomains": true,			 		\
			"Paths": ["/ipfs", "/ipns", "/api"]		\
		},											\
		"localhost": {								\
			"UseSubdomains": true,					\
			"InlineDNSLink": true,					\
			"Paths": ["/ipfs", "/ipns", "/api"]		\
		}											\
	}'
	# todo: SUBDOMAIN_GATEWAY_URL should be a cmd parameter
	./gateway-conformance test --json output.json --gateway-url ${GATEWAY_URL} --subdomain-url ${SUBDOMAIN_GATEWAY_URL} --specs +subdomain-gateway

test-kubo: provision-kubo
	GATEWAY_URL=http://127.0.0.1:8080 make _test

# tools
fixtures.car: gateway-conformance
	./gateway-conformance extract-fixtures --merged=true --dir=.

gateway-conformance:
	go build -o ./gateway-conformance ./cmd/gateway-conformance

_test: fixtures.car gateway-conformance
	./gateway-conformance test --json output.json --gateway-url ${GATEWAY_URL}

test-docker: fixtures.car gateway-conformance
	docker build -t gateway-conformance .
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" --network=host gateway-conformance test

output.xml: test-kubo
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" --entrypoint "/bin/bash" ghcr.io/pl-strflt/saxon:v1 -c """
		java -jar /opt/SaxonHE11-5J/saxon-he-11.5.jar -s:<(jq -s '.' output.json) -xsl:/etc/gotest.xsl -o:output.xml
	"""

output.html: output.xml
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" ghcr.io/pl-strflt/saxon:v1 -s:output.xml -xsl:/etc/junit-noframes-saxon.xsl -o:output.html
	open ./output.html

.PHONY: gateway-conformance
