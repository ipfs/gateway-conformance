package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	"github.com/ipfs/gateway-conformance/tooling/ipns"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
	. "github.com/ipfs/gateway-conformance/tooling/tmpl"
)

func TestGatewayCache(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0116/gateway-cache.car")

	// TODO: Add request chaining support to the test framework and enable the etag tests
	// https://specs.ipfs.tech/http-gateways/path-gateway/#etag-response-header
	// var etag string

	tests := SugarTests{
		{
			Name: "GET for /ipfs/ unixfs dir listing succeeds",
			Request: Request().
				Path("ipfs/{{CID}}/root2/root3/", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						IsEmpty(),
					Header("X-Ipfs-Path").
						Equals("/ipfs/{{CID}}/root2/root3/", fixture.MustGetCid()),
					Header("X-Ipfs-Roots").
						Equals("{{CID1}},{{CID2}},{{CID3}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3")),
					Header("Etag").
						Matches("DirIndex-.*_CID-{{}}", fixture.MustGetCid("root2", "root3")),
				),
		},
		{
			Name: "GET for /ipfs/ unixfs dir with index.html succeeds",
			Request: Request().
				Path("ipfs/{{}}/root2/root3/root4/", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						Equals("public, max-age=29030400, immutable"),
					Header("X-Ipfs-Path").
						Equals("/ipfs/{{CID}}/root2/root3/root4/", fixture.MustGetCid()),
					Header("X-Ipfs-Roots").
						Equals("{{CID1}},{{CID2}},{{CID3}},{{CID4}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4")),
					Header("Etag").
						Equals("\"{{CID}}\"", fixture.MustGetCid("root2", "root3", "root4")),
				),
		},
		{
			Name: "GET for /ipfs/ unixfs file succeeds",
			Request: Request().
				Path("ipfs/{{CID}}/root2/root3/root4/index.html", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						Equals("public, max-age=29030400, immutable"),
					Header("X-Ipfs-Path").
						Equals("/ipfs/{{CID}}/root2/root3/root4/index.html", fixture.MustGetCid()),
					Header("X-Ipfs-Roots").
						Equals("{{}},{{}},{{}},{{}},{{}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4"), fixture.MustGetCid("root2", "root3", "root4", "index.html")),
					Header("Etag").
						Equals("\"{{}}\"", fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				),
		},
		{
			Name: "GET for /ipfs/ unixfs dir as DAG-JSON succeeds",
			Request: Request().
				Path("ipfs/{{}}/root2/root3/root4/?format=dag-json", fixture.MustGetCid()),
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
				Path("ipfs/{{}}/root2/root3/root4/?format=json", fixture.MustGetCid()),
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
				Path("ipfs/{{}}/root2/root3/root4/?format=json", fixture.MustGetCid()).
				Headers(
					Header("Cache-Control", "only-if-cached"),
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
					Header("Cache-Control", "only-if-cached"),
				).
				Method("HEAD"),
			Response: Expect().
				Status(412),
		},
		{
			Name: "GET for /ipfs/ with only-if-cached succeeds when in local datastore",
			Request: Request().
				Path("ipfs/{{}}/root2/root3/root4/?format=json", fixture.MustGetCid()).
				Headers(
					Header("Cache-Control", "only-if-cached"),
				),
			Response: Expect().
				Status(200),
		},
		{
			Name: "GET for /ipfs/ with only-if-cached fails when not in local datastore",
			Request: Request().
				Path("ipfs/QmYzfKSE55XCjD1MW128RfciAf2DViABhEiXfgVFMabSjN").
				Headers(
					Header("Cache-Control", "only-if-cached"),
				),
			Response: Expect().
				Status(412),
		},
		{
			Name: "GET for /ipfs/ file with matching Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/{{}}/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", Templated("\"{{}}\"", fixture.MustGetCid("root2", "root3", "root4", "index.html"))),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ dir with index.html file with matching Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/{{}}/root2/root3/root4/", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", Templated("\"{{}}\"", fixture.MustGetCid("root2", "root3", "root4"))),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ file with matching third Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/{{}}/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", Templated("\"fakeEtag1\", \"fakeEtag2\", \"{{}}\"", fixture.MustGetCid("root2", "root3", "root4", "index.html"))),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ file with matching weak Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/{{}}/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", Templated("W/\"{{}}\"", fixture.MustGetCid("root2", "root3", "root4", "index.html"))),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ file with wildcard Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/{{}}/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", "*"),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ dir listing with matching weak Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/{{}}/root2/root3/", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", Templated("W/\"{{}}\"", fixture.MustGetCid("root2", "root3"))),
				),
			Response: Expect().
				Status(304),
		},
		// The tests below require `etag` to be set.
		/*
			{
				Name: "GET for /ipfs/ dir listing with matching strong Etag in If-None-Match returns 304 Not Modified",
				Request: Request().
					Path("ipfs/{{}}/root2/root3/", fixture.MustGetCid()).
					Headers(
						Header("If-None-Match", fmt.Sprintf("\"{{}}\"", etag)),
					),
				Response: Expect().
					Status(304),
			},
			{
				Name: "GET for /ipfs/ dir listing with matching strong Etag in If-None-Match returns 304 Not Modified",
				Request: Request().
					Path("ipfs/{{}}/root2/root3/", fixture.MustGetCid()).
					Headers(
						Header("If-None-Match", fmt.Sprintf("W/\"{{}}\"", etag)),
					),
				Response: Expect().
					Status(304),
			},
		*/
	}

	Run(t, tests)
}

func TestGatewayCacheWithIPNS(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0116/gateway-cache.car")
	ipns := ipns.MustOpenIPNSRecordWithKey("t0116/k51qzi5uqu5dlxdsdu5fpuu7h69wu4ohp32iwm9pdt9nq3y5rpn3ln9j12zfhe.ipns-record")
	ipnsKey := ipns.Key()

	tests := SugarTests{
		{
			Name: "GET for /ipns/ unixfs dir listing succeeds",
			Request: Request().
				Path("ipns/{{key}}/root2/root3/", ipnsKey),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						IsEmpty(),
					Header("X-Ipfs-Path").
						Equals("/ipns/{{key}}/root2/root3/", ipnsKey),
					Header("X-Ipfs-Roots").
						Equals("{{cid1}},{{cid2}},{{cid3}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3")),
					Header("Etag").
						Matches("DirIndex-.*_CID-{{cid}}", fixture.MustGetCid("root2", "root3")),
				),
		},
		{
			Name: "GET for /ipns/ unixfs dir with index.html succeeds",
			Request: Request().
				Path("ipns/{{key}}/root2/root3/root4/", ipnsKey),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						IsEmpty(),
					Header("X-Ipfs-Path").
						Equals("/ipns/{{key}}/root2/root3/root4/", ipnsKey),
					Header("X-Ipfs-Roots").
						Equals("{{cid1}},{{cid2}},{{cid3}},{{cid4}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4")),
					Header("Etag").
						Matches(`"{{cid}}"`, fixture.MustGetCid("root2", "root3", "root4")),
				),
		},
		{
			Name: "GET for /ipns/ unixfs file succeeds",
			Request: Request().
				Path("ipns/{{key}}/root2/root3/root4/index.html", ipnsKey),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						IsEmpty(),
					Header("X-Ipfs-Path").
						Equals("/ipns/{{key}}/root2/root3/root4/index.html", ipnsKey),
					Header("X-Ipfs-Roots").
						Equals("{{cid1}},{{cid2}},{{cid3}},{{cid4}},{{cid5}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4"), fixture.MustGetCid("root2", "root3", "root4", "index.html")),
					Header("Etag").
						Equals(`"{{cid}}"`, fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				),
		},
		{
			Name: "GET for /ipns/ unixfs dir as DAG-JSON succeeds",
			Request: Request().
				Path("ipns/{{key}}/root2/root3/root4/", ipnsKey).
				Query("format", "dag-json"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						IsEmpty(),
				),
		},
		{
			Name: "GET for /ipns/ unixfs dir as JSON succeeds",
			Request: Request().
				Path("ipns/{{key}}/root2/root3/root4/", ipnsKey).
				Query("format", "json"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						IsEmpty(),
				),
		},
		{
			Name: "GET for /ipns/ file with matching Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipns/{{key}}/root2/root3/root4/index.html", ipnsKey).
				Headers(
					Header("If-None-Match", Templated(`"{{cid}}"`, fixture.MustGetCid("root2", "root3", "root4", "index.html"))),
				),
			Response: Expect().
				Status(304),
		},
	}

	RunIfSpecsAreEnabled(
		t,
		tests,
		specs.IPNSResolver,
	)
}
