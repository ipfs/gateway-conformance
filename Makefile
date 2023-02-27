provision-cargateway:
	# cd go-libipfs/examples/car && go install
	car -c ./fixtures.car &

test-cargateway: provision-cargateway
	GATEWAY_URL=http://localhost:8040 make test

provision-kubo:
	find ./fixtures -name '*.car' -exec ipfs dag import {} \;

test-kubo: provision-kubo
	GATEWAY_URL=http://localhost:8080 make test

# tools
fixtures.car: generate
	./generate

generate: ./generate_fixture.go
	go build -o generate generate_fixture.go

test: fixtures.car
	# go install gotest.tools/gotestsum@latest
	gotestsum --junitfile output.xml