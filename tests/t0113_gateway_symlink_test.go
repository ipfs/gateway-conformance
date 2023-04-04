package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	"github.com/ipfs/gateway-conformance/tooling/test"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewaySymlink(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0113-gateway-symlink.car")

	tests := SugarTests{
		{
			Name: "Test the directory listing",
			Request: Request().
				Path("ipfs/%s?format=raw", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Body(fixture.MustGetRawData()),
		},
		{
			Name: "Test the symlink",
			Request: Request().
				Path("ipfs/%s/bar", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Bytes("foo"),
		},
	}

	test.Run(t, tests)
}
