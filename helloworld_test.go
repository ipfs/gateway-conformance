package main

import (
	"fmt"
	"regexp"
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
			Headers: map[string]interface{}{
				"Content-Type": test.WithHint[*regexp.Regexp]{
					Value: regexp.MustCompile("application.vnd.ipld.raw"),
					Hint:  "https://www.iana.org/assignments/media-types/application/vnd.ipld.raw",
				},
				"Content-Length": fmt.Sprintf("%d", len(car.GetRawBlock("fixtures/dir.car", "/dir/ascii.txt"))),
			},
		},
	},
}

func TestRawBlockSupport(t *testing.T) {
	test.Run(t, tests)
}
