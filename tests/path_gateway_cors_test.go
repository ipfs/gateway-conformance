package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestCors(t *testing.T) {
	cidHello := "bafkqabtimvwgy3yk" // hello

	tests := SugarTests{
		{
			Name: "GET Responses from Gateway should include CORS headers allowing JS from other origins to read the data cross-origin.",
			Request: Request().
				Path("/ipfs/{{CID}}/", cidHello),
			Response: Expect().
				Headers(
					Header("Access-Control-Allow-Origin").Equals("*"),
					Header("Access-Control-Allow-Methods").Has("GET", "HEAD", "OPTIONS"),
					Header("Access-Control-Allow-Headers").Has("Content-Type", "Range", "User-Agent", "X-Requested-With"),
					Header("Access-Control-Expose-Headers").Has(
						"Content-Range",
						"Content-Length",
						"X-Ipfs-Path",
						"X-Ipfs-Roots",
						"X-Chunked-Output",
						"X-Stream-Output",
					),
				),
		},
		{
			Name: "OPTIONS to Gateway succeeds",
			Request: Request().
				Method("OPTIONS").
				Path("/ipfs/{{CID}}/", cidHello),
			Response: Expect().
				Headers(
					Header("Access-Control-Allow-Origin").Equals("*"),
					Header("Access-Control-Allow-Methods").Has("GET", "HEAD", "OPTIONS"),
					Header("Access-Control-Allow-Headers").Has("Content-Type", "Range", "User-Agent", "X-Requested-With"),
					Header("Access-Control-Expose-Headers").Has(
						"Content-Range",
						"Content-Length",
						"X-Ipfs-Path",
						"X-Ipfs-Roots",
						"X-Chunked-Output",
						"X-Stream-Output",
					),
				),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS)
}
