package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	"github.com/ipfs/gateway-conformance/tooling/test"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewaySymlink(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0113-gateway-symlink.car")

	tests := []CTest{
		{
			Name: "Test the directory listing",
			Request: Request().
				Path("ipfs/%s?format=raw", fixture.MustGetCid()).Request(),
			Response: Expect().
				Status(200).
				Body(fixture.MustGetRawData()).
				Response(),
		},
		{
			Name: "Test the symlink",
			Request: Request().
				Path("ipfs/%s/bar", fixture.MustGetCid()).Request(),
			Response: Expect().
				Status(200).
				Bytes("foo").
				Response(),
		},
	}

	test.Run(t, tests)
}
