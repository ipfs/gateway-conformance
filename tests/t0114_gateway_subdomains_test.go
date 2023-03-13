package tests

import (
	"fmt"
	"testing"

	. "github.com/ipfs/gateway-conformance/tooling/check"
	. "github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewaySubdomains(t *testing.T) {
	// fixture := car.MustOpenUnixfsCar("t0114-gateway_subdomains")
	// TODO: extract the CID correctly.
	// TODO: how to extract CIDv1, CIDv0, and DIR_CID from the fixture?
	CIDv1 := "bafkreicysg23kiwv34eg2d7qweipxwosdo2py4ldv42nbauguluen5v6am"
	DIR_CID := "bafybeiht6dtwk3les7vqm6ibpvz6qpohidvlshsfyr7l5mpysdw2vmbbhe"

	tests := []CTest{
		{
			Name: "request for {gateway}/ipfs/{CIDv1} returns HTTP 301 Moved Permanently",
			Request: CRequest{
				DoNotFollowRedirects: true,
				Url:                  fmt.Sprintf("%s/ipfs/%s", SubdomainGatewayUrl, CIDv1),
			},
			Response: CResponse{
				StatusCode: 301,
				Headers: map[string]interface{}{
					"Location": Contains("%s://%s.ipfs.%s", SubdomainGatewayScheme, CIDv1, SubdomainGatewayHost),
				},
			},
		},
		{
			Name: "request for {gateway}/ipfs/{CIDv1} returns HTTP 301 Moved Permanently (sugar)",
			Request: Request().
				URL("http://example.com/ipfs/%s", CIDv1).
				DoNotFollowRedirects().
				Request(),
			Response: Expect().
				Status(301).
				Headers(
					Header("Location").
						// TODO: this works only because we use example.com in our tests.
						// It should be:
						// Contains("%s://%s.ipfs.%s", SubdomainGatewayScheme, CIDv1, SubdomainGatewayHost)
						// I am trying to avoid this syntax.
						// The other option is to force the tested gateway to use example.com.
						Contains("http://%s.ipfs.example.com", CIDv1),
				).
				Response(),
		},
		{
			Name: "request for {cid}.ipfs.localhost/api returns data if present on the content root",
			Request: CRequest{
				Url: fmt.Sprintf("%s://%s.ipfs.%s/api/file.txt", SubdomainGatewayScheme, DIR_CID, SubdomainGatewayHost),
			},
			Response: CResponse{
				Body: Contains("I am a txt file"),
			},
		},
		{
			Name: "request for {cid}.ipfs.localhost/api returns data if present on the content root (sugar)",
			Request: Request().
				URL("http://%s.ipfs.example.com/api/file.txt", DIR_CID).
				Request(),
			Response: Expect().
				Status(200).
				Body("I am a txt file\n").
				Response(),
		},
	}

	if SubdomainGateway.IsEnabled() {
		Run(t, tests)
	}
}
