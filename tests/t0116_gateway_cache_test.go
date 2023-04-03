package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

/* These tests do not cover the following:
** - /ipns/ paths
** - "If-None-Match" header handling for strong ETags for dir listings (the ones with xxhash)
 */
func TestGatewayCache(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0116-gateway-cache.car")

	tests := SugarTests{
		{
			Name: "GET for /ipfs/ unixfs dir listing succeeds",
			Request: Request().
				Path("ipfs/%s/root2/root3/", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						IsEmpty(),
					Header("X-Ipfs-Path").
						Equals("/ipfs/%s/root2/root3/", fixture.MustGetCid()),
					Header("X-Ipfs-Roots").
						Equals("%s,%s,%s", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3")),
					Header("Etag").
						Matches("DirIndex-.*_CID-%s", fixture.MustGetCid("root2", "root3")),
				),
		},
		{
			Name: "GET for /ipfs/ unixfs dir with index.html succeeds",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						Equals("public, max-age=29030400, immutable"),
					Header("X-Ipfs-Path").
						Equals("/ipfs/%s/root2/root3/root4/", fixture.MustGetCid()),
					Header("X-Ipfs-Roots").
						Equals("%s,%s,%s,%s", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4")),
					Header("Etag").
						Equals("\"%s\"", fixture.MustGetCid("root2", "root3", "root4")),
				),
		},
		{
			Name: "GET for /ipfs/ unixfs file succeeds",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						Equals("public, max-age=29030400, immutable"),
					Header("X-Ipfs-Path").
						Equals("/ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()),
					Header("X-Ipfs-Roots").
						Equals("%s,%s,%s,%s,%s", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4"), fixture.MustGetCid("root2", "root3", "root4", "index.html")),
					Header("Etag").
						Equals("\"%s\"", fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				),
		},
		{
			Name: "GET for /ipfs/ unixfs dir as DAG-JSON succeeds",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/?format=dag-json", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						Equals("public, max-age=29030400, immutable"),
				),
		},
		{
			Name: "GET for /ipfs/ unixfs dir as JSON succeeds",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/?format=json", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						Equals("public, max-age=29030400, immutable"),
				),
		},
		{
			Name: "HEAD for /ipfs/ with only-if-cached succeeds when in local datastore",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/?format=json", fixture.MustGetCid()).
				Headers(
					Header("Cache-Control").
						Equals("only-if-cached"),
				).
				Method("HEAD"),
			Response: Expect().
				Status(200),
		},
		{
			Name: "HEAD for /ipfs/ with only-if-cached fails when not in local datastore",
			Request: Request().
				Path("ipfs/QmYzfKSE55XCjD1MW128RfciAf2DViABhEiXfgVFMabSjN").
				Headers(
					Header("Cache-Control").
						Equals("only-if-cached"),
				).
				Method("HEAD"),
			Response: Expect().
				Status(412),
		},
		{
			Name: "GET for /ipfs/ with only-if-cached succeeds when in local datastore",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/?format=json", fixture.MustGetCid()).
				Headers(
					Header("Cache-Control").
						Equals("only-if-cached"),
				),
			Response: Expect().
				Status(200),
		},
		{
			Name: "GET for /ipfs/ with only-if-cached fails when not in local datastore",
			Request: Request().
				Path("ipfs/QmYzfKSE55XCjD1MW128RfciAf2DViABhEiXfgVFMabSjN").
				Headers(
					Header("Cache-Control").
						Equals("only-if-cached"),
				),
			Response: Expect().
				Status(412),
		},
		{
			Name: "GET for /ipfs/ file with matching Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match").
						Equals("\"%s\"", fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ dir with index.html file with matching Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match").
						Equals("\"%s\"", fixture.MustGetCid("root2", "root3", "root4")),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ file with matching third Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match").
						Equals("\"fakeEtag1\", \"fakeEtag2\", \"%s\"", fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ file with matching weak Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match").
						Equals("W/\"%s\"", fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ file with wildcard Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match").
						Equals("*"),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ dir listing with matching weak Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/%s/root2/root3/", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match").
						Equals("W/\"%s\"", fixture.MustGetCid("root2", "root3")),
				),
			Response: Expect().
				Status(304),
		},
	}.Build()

	Run(t, tests)
}
