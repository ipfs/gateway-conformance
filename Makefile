provision-cargateway: ./fixtures.car
	# cd go-libipfs/examples/car && go install
	car -c ./fixtures.car &

test-cargateway: provision-cargateway
	GATEWAY_URL=http://127.0.0.1:8040 make _test

provision-kubo:
	find ./fixtures -name '*.car' -exec ipfs dag import {} \;

test-kubo: provision-kubo
	GATEWAY_URL=http://127.0.0.1:8080 make _test

merge-fixtures:
	go build -o merge-fixtures ./tooling/cmd/merge_fixtures.go

# tools
fixtures.car: entrypoint
	./entrypoint merge-fixtures ./fixtures.car

entrypoint:
	go build -o ./entrypoint ./entrypoint.go

_test: fixtures.car entrypoint
	./entrypoint test --json-output output.json --gateway-url ${GATEWAY_URL} --is-subdomain

test-docker: fixtures.car entrypoint
	docker build -t gway-test .
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" --network=host gway-test test

output.xml: test-kubo
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" --entrypoint "/bin/bash" ghcr.io/pl-strflt/saxon:v1 -c """
		java -jar /opt/SaxonHE11-5J/saxon-he-11.5.jar -s:<(jq -s '.' output.json) -xsl:/etc/gotest.xsl -o:output.xml
	"""

output.html: output.xml
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" ghcr.io/pl-strflt/saxon:v1 -s:output.xml -xsl:/etc/junit-noframes-saxon.xsl -o:output.html
	open ./output.html

.PHONY: entrypoint
