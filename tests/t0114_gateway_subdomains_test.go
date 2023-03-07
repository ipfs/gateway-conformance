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
	CIDv1 := "bafkreicysg23kiwv34eg2d7qweipxwosdo2py4ldv42nbauguluen5v6am"
	DIR_CID := "bafybeiht6dtwk3les7vqm6ibpvz6qpohidvlshsfyr7l5mpysdw2vmbbhe"

	tests := []CTest{
		{
			Name: "request for {gateway}/ipfs/{CIDv1} returns HTTP 301 Moved Permanently",
			Request: CRequest{
				DoNotFollowRedirects: true,
				RawURL: fmt.Sprintf("http://%s/ipfs/%s", SubdomainGatewayHost, CIDv1),
				// Url:                  fmt.Sprintf("ipfs/%s", CIDv1),
			},
			Response: CResponse{
				StatusCode: 301,
				Headers: map[string]interface{}{
					"Location": Contains("http://%s.ipfs.%s", CIDv1, SubdomainGatewayHost),
				},
			},
		},
		{
			Name: "request for {cid}.ipfs.localhost/api returns data if present on the content root",
			Request: CRequest{
				RawURL: fmt.Sprintf("http://%s.ipfs.%s/api/file.txt", DIR_CID, SubdomainGatewayHost),
			},
			Response: CResponse{
				Body: Contains("I am a txt file"),
			},
		},
	}

	if SubdomainGateway.IsEnabled() {
		Run(t, tests)
	}
}
