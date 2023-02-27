package main

import (
	"fmt"
	"testing"

	"github.com/ipfs/gateway-conformance/car"
	"github.com/ipfs/gateway-conformance/test"
)

func TestGatewaySymlink(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("fixtures/t0113-gateway-symlink.car")
	tests := map[string]test.Test{
		"Test the directory listing": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s?format=raw", fixture.MustGetCid()),
			},
			Response: test.Response{
				StatusCode: 200,
				Body:       fixture.MustGetRawData(),
			},
		},
		"Test the symlink": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/bar", fixture.MustGetCid()),
			},
			Response: test.Response{
				StatusCode: 200,
				Body:       []byte("foo"),
			},
		},
	}

	test.Run(t, tests)
}
