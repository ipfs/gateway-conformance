package main

import (
	"fmt"
	"testing"

	"github.com/ipfs/gateway-conformance/car"
	"github.com/ipfs/gateway-conformance/test"
)

var tests = map[string]test.Test{
	"GET with format=raw param returns a raw block": {
		Request: test.Request{
			Url: fmt.Sprintf("ipfs/%s/dir/ascii.txt?format=raw", car.GetCid("fixtures/dir.car", "/")),
		},
		Response: test.Response{
			StatusCode: 200,
			Body:       car.GetRawBlock("fixtures/dir.car", "/dir/ascii.txt"),
			Headers: test.Headers{
				"Content-Type": test.StringWithHint{
					Value: "application/vnd.ipld.raw",
					Hint:  "https://www.iana.org/assignments/media-types/application/vnd.ipld.raw",
				},
				"Content-Length": test.String(fmt.Sprintf("%d", len(car.GetRawBlock("fixtures/dir.car", "/dir/ascii.txt")))),
			},
		},
	},
}

func TestRawBlockSupport(t *testing.T) {
	test.Run(t, tests)
}
