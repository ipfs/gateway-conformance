provision-cargateway: ./fixtures.car
	# cd go-libipfs/examples/car && go install
	car -c ./fixtures.car &

test-cargateway: provision-cargateway
	GATEWAY_URL=http://127.0.0.1:8040 make test

provision-kubo:
	find ./fixtures -name '*.car' -exec ipfs dag import {} \;

test-kubo: provision-kubo
	GATEWAY_URL=http://127.0.0.1:8080 make test

test-kubo-subdomains: provision-kubo
	ipfs config --json Gateway.PublicGateways '{	\
	"example.com": {								\
		"UseSubdomains": true,			 			\
		"Paths": ["/ipfs", "/ipns", "/api"]			\
	}												\
	}'
	SUBDOMAIN_GATEWAY_URL=http://example.com:8080 GOTAGS=test_subdomains make test-kubo


merge-fixtures:
	go build -o merge-fixtures ./tooling/cmd/merge_fixtures.go

# tools
fixtures.car: merge-fixtures
	./merge-fixtures ./fixtures.car

test: fixtures.car
	# go install gotest.tools/gotestsum@latest
	- gotestsum --format testname --junitfile output.xml ./tests -tags "${GOTAGS}"

output.xml: test-kubo

output.html: output.xml
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" ghcr.io/pl-strflt/junit-xml-to-html:latest no-frames ./output.xml ./output.html
	open ./output.html

.PHONY: merge-fixtures