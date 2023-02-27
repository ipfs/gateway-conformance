package main

import (
	"fmt"
	"testing"

	"github.com/ipfs/gateway-conformance/car"
	"github.com/ipfs/gateway-conformance/test"
)

func TestGatewayCar(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("fixtures/t0118-test-dag.car")

	// CAR stream is not deterministic, as blocks can arrive in random order,
	// but if we have a small file that fits into a single block, and export its CID
	// we will get a CAR that is a deterministic array of bytes.
	tests := map[string]test.Test{
		"GET response for application/vnd.ipld.car": {
			// Test between l85 and l112
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/subdir/ascii.txt", fixture.MustGetCid()),
				Headers: map[string]string{
					"Accept": "application/vnd.ipld.car",
				},
			},
			Response: test.Response{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"Content-Type": test.HeaderContains(
						"Expected content type to be application/vnd.ipld.car",
						"application/vnd.ipld.car",
					),
					"Content-Length": test.H(
						"CAR is streamed, gateway may not have the entire thing, unable to calculate total size",
						""),
					"Content-Disposition": test.HeaderContains(
						"Expected content disposition to be attachment; filename=\"<cid>.car\"",
						fmt.Sprintf("attachment\\; filename=\"%s.car\"", fixture.MustGetCid("subdir", "ascii.txt"))),
					"X-Content-Type-Options": "nosniff",
					"Accept-Ranges": test.H(
						"CAR is streamed, gateway may not have the entire thing, unable to support range-requests. Partial downloads and resumes should be handled using IPLD selectors: https://github.com/ipfs/go-ipfs/issues/8769",
						"none",
					),
				},
			},
		},
	}

	test.Run(t, tests)
}
