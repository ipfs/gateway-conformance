package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewayCar(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0118-test-dag.car")

	tests := SugarTests{
		{
			Name: "GET response for application/vnd.ipld.car",
			Hint: `
				CAR stream is not deterministic, as blocks can arrive in random order,
				but if we have a small file that fits into a single block, and export its CID
				we will get a CAR that is a deterministic array of bytes.
			`,
			Request: Request().
				Path("ipfs/%s/subdir/ascii.txt", fixture.MustGetCid()).
				Headers(
					Header("Accept", "application/vnd.ipld.car"),
				),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").
						Hint("Expected content type to be application/vnd.ipld.car").
						Contains("application/vnd.ipld.car"),
					Header("Content-Length").
						Hint("CAR is streamed, gateway may not have the entire thing, unable to calculate total size").
						IsEmpty(),
					Header("Content-Disposition").
						Hint("Expected content disposition to be attachment; filename=\"<cid>.car\"").
						Contains("attachment; filename=\"%s.car\"", fixture.MustGetCid("subdir", "ascii.txt")),
					Header("X-Content-Type-Options").
						Hint("CAR is streamed, gateway may not have the entire thing, unable to calculate total size").
						Equals("nosniff"),
					Header("Accept-Ranges").
						Hint("CAR is streamed, gateway may not have the entire thing, unable to support range-requests. Partial downloads and resumes should be handled using IPLD selectors: https://github.com/ipfs/go-ipfs/issues/8769").
						Equals("none"),
				).Body(
					IsCar().
					 	HasRoot(fixture.MustGetCid("subdir", "ascii.txt")).
						HasBlock(fixture.MustGetCid("subdir", "ascii.txt")).
						Exactly().
						InThatOrder(),	
				),
		},
	}

	Run(t, tests)
}
