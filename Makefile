provision-cargateway: ./fixtures.car
	# cd go-libipfs/examples/car && go install
	car -c ./fixtures.car &

test-cargateway: provision-cargateway
	GATEWAY_URL=http://localhost:8040 make test

provision-kubo:
	find ./fixtures -name '*.car' -exec ipfs dag import {} \;

test-kubo: provision-kubo
	GATEWAY_URL=http://localhost:8080 make test

merge-fixtures:
	go build -o merge-fixtures ./cmd/merge_fixtures.go

# tools
fixtures.car: merge-fixtures
	./merge-fixtures

test: fixtures.car
	# go install gotest.tools/gotestsum@latest
	- gotestsum --junitfile output.xml

output.xml: test-kubo

output.html: output.xml
	docker run --rm -v "${PWD}:/workspace" -w "/workspace" ghcr.io/pl-strflt/junit-xml-to-html:latest no-frames ./output.xml ./output.html
	open ./output.html

.PHONY: merge-fixtures