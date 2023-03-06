//go:build test_subdomains
// +build test_subdomains

package main

import (
	"fmt"
	"testing"

	. "github.com/ipfs/gateway-conformance/tooling/check"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewaySubdomains(t *testing.T) {
	// fixture := car.MustOpenUnixfsCar("t0114-gateway_subdomains")

	CID_VAL := "hello"
	CIDv1 := "bafkreicysg23kiwv34eg2d7qweipxwosdo2py4ldv42nbauguluen5v6am"
	// CIDv0 := "QmZULkCELmmk5XNfCgTnCyFgAVxBRBXyDHGGMVoLFLiXEN"
	// // CIDv0to1 is necessary because raw-leaves are enabled by default during
	// // "ipfs add" with CIDv1 and disabled with CIDv0
	// CIDv0to1 := "bafybeiffndsajwhk3lwjewwdxqntmjm4b5wxaaanokonsggenkbw6slwk4"
	// CIDv1_TOO_LONG := "bafkrgqhhyivzstcz3hhswshfjgy6ertgmnqeleynhwt4dlfsthi4hn7zgh4uvlsb5xncykzapi3ocd4lzogukir6ksdy6wzrnz6ohnv4aglcs"
	DIR_CID := "bafybeiht6dtwk3les7vqm6ibpvz6qpohidvlshsfyr7l5mpysdw2vmbbhe" // ./testdirlisting

	tests := []CTest{
		{
			Name: "request for 127.0.0.1/ipfs/{CID} stays on path",
			Request: CRequest{
				Url: fmt.Sprintf("ipfs/%s", CIDv1),
			},
			Response: CResponse{
				StatusCode: 200,
				Body:       Contains(CID_VAL),
			},
		},
		{
			Name: "request for localhost/ipfs/{CIDv1} returns HTTP 301 Moved Permanently",
			Request: CRequest{
				DoNotFollowRedirects: true,
				Url:                  fmt.Sprintf("/ipfs/%s", CIDv1),
			},
			Response: CResponse{
				StatusCode: 301,
				Headers: map[string]interface{}{
					"Location": Contains("http://%s.ipfs.localhost:8080", CIDv1),
				},
			},
		},
		{
			Name: "request for {cid}.ipfs.localhost/api returns data if present on the content root",
			Request: CRequest{
				RawURL: fmt.Sprintf("http://%s.ipfs.localhost:8080/api/file.txt", DIR_CID),
			},
			Response: CResponse{
				Body: Contains("I am a txt file"),
			},
		},
	}

	Run(t, tests)
}
