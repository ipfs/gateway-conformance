fixtures.car: generate
	./generate

generate: ./generate_fixture.go
	go build -p generate generate_fixtures.go
