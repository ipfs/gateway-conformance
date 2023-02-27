package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/ipfs/gateway-conformance/car"
	"github.com/ipfs/gateway-conformance/test"
)

func TestGatewayCache(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("fixtures/t0116-gateway-cache.car")
	tests := map[string]test.Test{
		"GET for /ipfs/ unixfs dir listing succeeds": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/", fixture.MustGetCid()),
			},
			Response: test.Response{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"Cache-Control": "",
					"X-Ipfs-Path":   fmt.Sprintf("/ipfs/%s/root2/root3/", fixture.MustGetCid()),
					"X-Ipfs-Roots":  fmt.Sprintf("%s,%s,%s", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3")),
					"Etag":          regexp.MustCompile(fmt.Sprintf("DirIndex-.*_CID-%s", fixture.MustGetCid("root2", "root3"))),
				},
			},
		},
		"GET for /ipfs/ unixfs dir with index.html succeeds": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/", fixture.MustGetCid()),
			},
			Response: test.Response{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"Cache-Control": "public, max-age=29030400, immutable",
					"X-Ipfs-Path":   fmt.Sprintf("/ipfs/%s/root2/root3/root4/", fixture.MustGetCid()),
					"X-Ipfs-Roots":  fmt.Sprintf("%s,%s,%s,%s", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4")),
					"Etag":          fmt.Sprintf("\"%s\"", fixture.MustGetCid("root2", "root3", "root4")),
				},
			},
		},
		"GET for /ipfs/ unixfs file succeeds": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()),
			},
			Response: test.Response{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"Cache-Control": "public, max-age=29030400, immutable",
					"X-Ipfs-Path":   fmt.Sprintf("/ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()),
					"X-Ipfs-Roots":  fmt.Sprintf("%s,%s,%s,%s,%s", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4"), fixture.MustGetCid("root2", "root3", "root4", "index.html")),
					"Etag":          fmt.Sprintf("\"%s\"", fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				},
			},
		},
		"GET for /ipfs/ unixfs dir as DAG-JSON succeeds": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/?format=dag-json", fixture.MustGetCid()),
			},
			Response: test.Response{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"Cache-Control": "public, max-age=29030400, immutable",
				},
			},
		},
		"GET for /ipfs/ unixfs dir as JSON succeeds": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/?format=json", fixture.MustGetCid()),
			},
			Response: test.Response{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"Cache-Control": "public, max-age=29030400, immutable",
				},
			},
		},
		"HEAD for /ipfs/ with only-if-cached succeeds when in local datastore": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/?format=json", fixture.MustGetCid()),
				Headers: map[string]string{
					"Cache-Control": "only-if-cached",
				},
				Method: "HEAD",
			},
			Response: test.Response{
				StatusCode: 200,
			},
		},
		"HEAD for /ipfs/ with only-if-cached fails when not in local datastore": {
			Request: test.Request{
				Url: "ipfs/QmYzfKSE55XCjD1MW128RfciAf2DViABhEiXfgVFMabSjN",
				Headers: map[string]string{
					"Cache-Control": "only-if-cached",
				},
				Method: "HEAD",
			},
			Response: test.Response{
				StatusCode: 412,
			},
		},
		"GET for /ipfs/ with only-if-cached succeeds when in local datastore": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/?format=json", fixture.MustGetCid()),
				Headers: map[string]string{
					"Cache-Control": "only-if-cached",
				},
			},
			Response: test.Response{
				StatusCode: 200,
			},
		},
		"GET for /ipfs/ with only-if-cached fails when not in local datastore": {
			Request: test.Request{
				Url: "ipfs/QmYzfKSE55XCjD1MW128RfciAf2DViABhEiXfgVFMabSjN",
				Headers: map[string]string{
					"Cache-Control": "only-if-cached",
				},
			},
			Response: test.Response{
				StatusCode: 412,
			},
		},
		"GET for /ipfs/ file with matching Etag in If-None-Match returns 304 Not Modified": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()),
				Headers: map[string]string{
					"If-None-Match": fmt.Sprintf("\"%s\"", fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				},
			},
			Response: test.Response{
				StatusCode: 304,
			},
		},
		"GET for /ipfs/ dir with index.html file with matching Etag in If-None-Match returns 304 Not Modified": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/", fixture.MustGetCid()),
				Headers: map[string]string{
					"If-None-Match": fmt.Sprintf("\"%s\"", fixture.MustGetCid("root2", "root3", "root4")),
				},
			},
			Response: test.Response{
				StatusCode: 304,
			},
		},
		"GET for /ipfs/ file with matching third Etag in If-None-Match returns 304 Not Modified": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()),
				Headers: map[string]string{
					"If-None-Match": fmt.Sprintf("\"fakeEtag1\", \"fakeEtag2\", \"%s\"", fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				},
			},
			Response: test.Response{
				StatusCode: 304,
			},
		},
		"GET for /ipfs/ file with matching weak Etag in If-None-Match returns 304 Not Modified": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()),
				Headers: map[string]string{
					"If-None-Match": fmt.Sprintf("W/\"%s\"", fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				},
			},
			Response: test.Response{
				StatusCode: 304,
			},
		},
		"GET for /ipfs/ file with wildcard Etag in If-None-Match returns 304 Not Modified": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()),
				Headers: map[string]string{
					"If-None-Match": "*",
				},
			},
			Response: test.Response{
				StatusCode: 304,
			},
		},
		"GET for /ipfs/ dir listing with matching weak Etag in If-None-Match returns 304 Not Modified": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/root2/root3/", fixture.MustGetCid()),
				Headers: map[string]string{
					"If-None-Match": fmt.Sprintf("W/\"%s\"", fixture.MustGetCid("root2", "root3")),
				},
			},
			Response: test.Response{
				StatusCode: 304,
			},
		},
	}

	test.Run(t, tests)
}
