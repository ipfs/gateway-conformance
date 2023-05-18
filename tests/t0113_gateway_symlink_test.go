package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/test"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewaySymlink(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0113-gateway-symlink.car")
	rootDirCID := fixture.MustGetCid()

	tests := SugarTests{
		{
			Name: "Test the directory listing",
			Request: Request().
				Path("ipfs/{{CID}}/", rootDirCID),
			Response: Expect().
				Body(
					And(
						Contains(">foo<"),
						Contains(">bar<"),
					),
				),
		},
		{
			Name: "Test the directory raw query",
			Request: Request().
				Path("ipfs/{{CID}}", rootDirCID).
				Query("format", "raw"),
			Response: Expect().
				Status(200).
				Body(fixture.MustGetRawData()),
		},
		{
			Name: "Test the symlink",
			Request: Request().
				Path("ipfs/{{CID}}/bar", rootDirCID),
			Response: Expect().
				Status(200).
				Bytes("foo"),
		},
	}

	test.Run(t, tests)
}
