package main

import (
	"fmt"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

/* These tests do not cover the following:
** - /ipns/ paths
** - "If-None-Match" header handling for strong ETags for dir listings (the ones with xxhash)
 */
func TestGatewayCache(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0116-gateway-cache.car")

	tests := []CTest{
		{
			Name: "GET for /ipfs/ unixfs dir listing succeeds",
			Request: CRequest{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/", fixture.MustGetCid()),
			},
			Response: CResponse{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"Cache-Control": IsEmpty(),
					"X-Ipfs-Path":   IsEqual("/ipfs/%s/root2/root3/", fixture.MustGetCid()),
					"X-Ipfs-Roots":  IsEqual("%s,%s,%s", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3")),
					"Etag":          Matches("DirIndex-.*_CID-%s", fixture.MustGetCid("root2", "root3")),
				},
			},
		},
		{
			Name: "GET for /ipfs/ unixfs dir with index.html succeeds",
			Request: CRequest{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/", fixture.MustGetCid()),
			},
			Response: CResponse{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"Cache-Control": "public, max-age=29030400, immutable",
					"X-Ipfs-Path":   IsEqual("/ipfs/%s/root2/root3/root4/", fixture.MustGetCid()),
					"X-Ipfs-Roots":  IsEqual("%s,%s,%s,%s", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4")),
					"Etag":          IsEqual("\"%s\"", fixture.MustGetCid("root2", "root3", "root4")),
				},
			},
		},
		{
			Name: "GET for /ipfs/ unixfs file succeeds",
			Request: CRequest{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()),
			},
			Response: CResponse{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"Cache-Control": "public, max-age=29030400, immutable",
					"X-Ipfs-Path":   IsEqual("/ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()),
					"X-Ipfs-Roots":  IsEqual("%s,%s,%s,%s,%s", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4"), fixture.MustGetCid("root2", "root3", "root4", "index.html")),
					"Etag":          IsEqual("\"%s\"", fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				},
			},
		},
		{
			Name: "GET for /ipfs/ unixfs dir as DAG-JSON succeeds",
			Request: CRequest{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/?format=dag-json", fixture.MustGetCid()),
			},
			Response: CResponse{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"Cache-Control": "public, max-age=29030400, immutable",
				},
			},
		},
		{
			Name: "GET for /ipfs/ unixfs dir as JSON succeeds",
			Request: CRequest{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/?format=json", fixture.MustGetCid()),
			},
			Response: CResponse{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"Cache-Control": "public, max-age=29030400, immutable",
				},
			},
		},
		{
			Name: "HEAD for /ipfs/ with only-if-cached succeeds when in local datastore",
			Request: CRequest{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/?format=json", fixture.MustGetCid()),
				Headers: map[string]string{
					"Cache-Control": "only-if-cached",
				},
				Method: "HEAD",
			},
			Response: CResponse{
				StatusCode: 200,
			},
		},
		{
			Name: "HEAD for /ipfs/ with only-if-cached fails when not in local datastore",
			Request: CRequest{
				Url: "ipfs/QmYzfKSE55XCjD1MW128RfciAf2DViABhEiXfgVFMabSjN",
				Headers: map[string]string{
					"Cache-Control": "only-if-cached",
				},
				Method: "HEAD",
			},
			Response: CResponse{
				StatusCode: 412,
			},
		},
		{
			Name: "GET for /ipfs/ with only-if-cached succeeds when in local datastore",
			Request: CRequest{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/?format=json", fixture.MustGetCid()),
				Headers: map[string]string{
					"Cache-Control": "only-if-cached",
				},
			},
			Response: CResponse{
				StatusCode: 200,
			},
		},
		{
			Name: "GET for /ipfs/ with only-if-cached fails when not in local datastore",
			Request: CRequest{
				Url: "ipfs/QmYzfKSE55XCjD1MW128RfciAf2DViABhEiXfgVFMabSjN",
				Headers: map[string]string{
					"Cache-Control": "only-if-cached",
				},
			},
			Response: CResponse{
				StatusCode: 412,
			},
		},
		{
			Name: "GET for /ipfs/ file with matching Etag in If-None-Match returns 304 Not Modified",
			Request: CRequest{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()),
				Headers: map[string]string{
					"If-None-Match": fmt.Sprintf("\"%s\"", fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				},
			},
			Response: CResponse{
				StatusCode: 304,
			},
		},
		{
			Name: "GET for /ipfs/ dir with index.html file with matching Etag in If-None-Match returns 304 Not Modified",
			Request: CRequest{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/", fixture.MustGetCid()),
				Headers: map[string]string{
					"If-None-Match": fmt.Sprintf("\"%s\"", fixture.MustGetCid("root2", "root3", "root4")),
				},
			},
			Response: CResponse{
				StatusCode: 304,
			},
		},
		{
			Name: "GET for /ipfs/ file with matching third Etag in If-None-Match returns 304 Not Modified",
			Request: CRequest{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()),
				Headers: map[string]string{
					"If-None-Match": fmt.Sprintf("\"fakeEtag1\", \"fakeEtag2\", \"%s\"", fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				},
			},
			Response: CResponse{
				StatusCode: 304,
			},
		},
		{
			Name: "GET for /ipfs/ file with matching weak Etag in If-None-Match returns 304 Not Modified",
			Request: CRequest{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()),
				Headers: map[string]string{
					"If-None-Match": fmt.Sprintf("W/\"%s\"", fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				},
			},
			Response: CResponse{
				StatusCode: 304,
			},
		}, {
			Name: "GET for /ipfs/ file with wildcard Etag in If-None-Match returns 304 Not Modified",
			Request: CRequest{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()),
				Headers: map[string]string{
					"If-None-Match": "*",
				},
			},
			Response: CResponse{
				StatusCode: 304,
			},
		}, {
			Name: "GET for /ipfs/ dir listing with matching weak Etag in If-None-Match returns 304 Not Modified",
			Request: CRequest{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/", fixture.MustGetCid()),
				Headers: map[string]string{
					"If-None-Match": fmt.Sprintf("W/\"%s\"", fixture.MustGetCid("root2", "root3")),
				},
			},
			Response: CResponse{
				StatusCode: 304,
			},
		},
	}

	Run(t, tests)
}
