package tests

import (
	"strings"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling"
	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/ipns"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
	. "github.com/ipfs/gateway-conformance/tooling/tmpl"
)

func TestUnixFSDirectoryListing(t *testing.T) {
	tooling.LogTestGroup(t, GroupUnixFS)

	fixture := car.MustOpenUnixfsCar("dir_listing/fixtures.car")
	root := fixture.MustGetNode()
	file := fixture.MustGetNode("ą", "ę", "file-źł.txt")

	tests := SugarTests{
		{
			Name: "path gw: backlink on root CID should be hidden (TODO: cleanup Kubo-specifics)",
			Request: Request().
				Path("/ipfs/{{cid}}/", root.Cid()),
			Response: Expect().
				Body(
					And(
						Contains("Index of"),
						Not(Contains(`<a href="/ipfs/{{cid}}/">..</a>`, root.Cid())),
					)),
		},
		{
			Name: "path gw: redirect dir listing to URL with trailing slash",
			Request: Request().
				Path("/ipfs/{{cid}}/ą/ę", root.Cid()),
			Response: Expect().
				Status(301).
				Headers(
					Header("Location", `/ipfs/{{cid}}/%C4%85/%C4%99/`, root.Cid()),
				),
		},
		{
			Name: "path gw: dir listing HTML response (TODO: cleanup Kubo-specifics)",
			Request: Request().
				Path("/ipfs/{{cid}}/ą/ę/", root.Cid()),
			Response: Expect().
				Headers(
					Header("Etag").Contains(`"DirIndex-`),
				).
				BodyWithHint(`
				- should contain "Index of" (TODO:  kubo-specific)
				- Breadcrumbs should point at /ipfs namespace mounted at Origin root (TODO:  kubo-specific)
				- backlink on subdirectory should point at parent directory
				- name column should be a link to its content path
				- hash column should be a CID link with filename param
				`,
					And(
						Contains("Index of"),
						Contains(`/ipfs/<a href="/ipfs/{{cid}}">{{cid}}</a>/<a href="/ipfs/{{cid}}/%C4%85">ą</a>/<a href="/ipfs/{{cid}}/%C4%85/%C4%99">ę</a>`,
							root.Cid()),
						Contains(`<a href="/ipfs/{{cid}}/%C4%85/%C4%99/..">..</a>`, root.Cid()),
						Contains(`<a href="/ipfs/{{cid}}/%C4%85/%C4%99/file-%C5%BA%C5%82.txt">file-źł.txt</a>`, root.Cid()),
						Contains(`<a class="ipfs-hash" translate="no" href="/ipfs/{{cid}}?filename=file-%25C5%25BA%25C5%2582.txt">`, file.Cid())),
				),
		},
		{
			Name: "GET for /ipfs/cid/file UnixFS file that does not exist returns 404",
			Request: Request().
				Path("/ipfs/{{cid}}/i-do-not-exist", root.Cid()),
			Response: Expect().
				Status(404),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS)
}

func TestGatewayCache(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("gateway-cache/fixtures.car")

	tests := SugarTests{
		{
			Name: "GET for /ipfs/ unixfs dir listing succeeds",
			Hint: "UnixFS directory listings are generated HTML, which may change over time, and can't be cached forever. Still, should have a meaningful cache-control header.",
			Request: Request().
				Path("/ipfs/{{CID}}/root2/root3/", fixture.MustGetCid()),
			Response: AllOf(
				Expect().
					Status(200).
					Headers(
						Header("X-Ipfs-Path").
							Equals("/ipfs/{{CID}}/root2/root3/", fixture.MustGetCid()),
						Header("X-Ipfs-Roots").
							Equals("{{CID1}},{{CID2}},{{CID3}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3")),
						Header("Etag").
							Matches("DirIndex-.*_CID-{{cid}}", fixture.MustGetCid("root2", "root3")),
					),
				AnyOf(
					Expect().Headers(Header("Cache-Control").IsEmpty()),
					Expect().Headers(Header("Cache-Control").Equals("public, max-age=604800, stale-while-revalidate=2678400")),
				),
			),
		},
		{
			Name: "GET for /ipfs/ unixfs dir with index.html succeeds",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/", fixture.MustGetCid()),
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
						Equals(`"{{CID}}"`, fixture.MustGetCid("root2", "root3", "root4")),
				),
		},
		{
			Name: "GET for /ipfs/ unixfs file succeeds",
			Request: Request().
				Path("/ipfs/{{CID}}/root2/root3/root4/index.html", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						Equals("public, max-age=29030400, immutable"),
					Header("X-Ipfs-Path").
						Equals("/ipfs/{{CID}}/root2/root3/root4/index.html", fixture.MustGetCid()),
					Header("X-Ipfs-Roots").
						Equals("{{cid1}},{{cid2}},{{cid3}},{{cid4}},{{cid5}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4"), fixture.MustGetCid("root2", "root3", "root4", "index.html")),
					Header("Etag").
						Equals(`"{{cid}}"`, fixture.MustGetCid("root2", "root3", "root4", "index.html")),
				),
		},
		{
			Name: "GET for /ipfs/ unixfs dir as raw block succeeds",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/?format=raw", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						Equals("public, max-age=29030400, immutable"),
				).
				Body(
					IsEqualBytes([]byte{
						0x12, 0x34, 0x0a, 0x24, 0x01, 0x55, 0x12, 0x20, 0x58, 0x91, 0xb5, 0xb5,
						0x22, 0xd5, 0xdf, 0x08, 0x6d, 0x0f, 0xf0, 0xb1, 0x10, 0xfb, 0xd9, 0xd2,
						0x1b, 0xb4, 0xfc, 0x71, 0x63, 0xaf, 0x34, 0xd0, 0x82, 0x86, 0xa2, 0xe8,
						0x46, 0xf6, 0xbe, 0x03, 0x12, 0x0a, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2e,
						0x68, 0x74, 0x6d, 0x6c, 0x18, 0x06, 0x0a, 0x02, 0x08, 0x01,
					}),
				),
		},
		{
			Name: "HEAD for /ipfs/ with only-if-cached succeeds when in local datastore",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/?format=raw", fixture.MustGetCid()).
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
				Path("/ipfs/QmYzfKSE55XCjD1MW128RfciAf2DViABhEiXfgVFMabSjN").
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
				Path("/ipfs/{{cid}}/root2/root3/root4/?format=raw", fixture.MustGetCid()).
				Headers(
					Header("Cache-Control", "only-if-cached"),
				),
			Response: Expect().
				Status(200),
		},
		{
			Name: "GET for /ipfs/ with only-if-cached fails when not in local datastore",
			Request: Request().
				Path("/ipfs/QmYzfKSE55XCjD1MW128RfciAf2DViABhEiXfgVFMabSjN").
				Headers(
					Header("Cache-Control", "only-if-cached"),
				),
			Response: Expect().
				Status(412),
		},
		// ==========
		// # If-None-Match (return 304 Not Modified when client sends matching Etag they already have)
		// ==========
		{
			Name: "GET for /ipfs/ file with matching Etag in If-None-Match returns 304 Not Modified",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#if-none-match-request-header",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", Fmt(`"{{cid}}"`, fixture.MustGetCid("root2", "root3", "root4", "index.html"))),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ dir with index.html file with matching Etag in If-None-Match returns 304 Not Modified",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#if-none-match-request-header",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", Fmt(`"{{cid}}"`, fixture.MustGetCid("root2", "root3", "root4"))),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ file with matching third Etag in If-None-Match returns 304 Not Modified",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#if-none-match-request-header",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", Fmt(`"fakeEtag1", "fakeEtag2", "{{cid}}"`, fixture.MustGetCid("root2", "root3", "root4", "index.html"))),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ file with matching weak Etag in If-None-Match returns 304 Not Modified",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#if-none-match-request-header",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", Fmt(`W/"{{cid}}"`, fixture.MustGetCid("root2", "root3", "root4", "index.html"))),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ file with wildcard Etag in If-None-Match returns 304 Not Modified",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#if-none-match-request-header",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", "*"),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ dir listing with matching weak Etag in If-None-Match returns 304 Not Modified",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#if-none-match-request-header",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", Fmt(`W/"{{cid}}"`, fixture.MustGetCid("root2", "root3"))),
				),
			Response: Expect().
				Status(304),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS)

	// DirIndex etagDir is based on xxhash(./assets/dir-index-html), so we need to fetch it dynamically
	var etagDir string

	testsA := SugarTests{
		{
			Name: "DirIndex etag is based on xxhash(./assets/dir-index-html), so we need to fetch it dynamically",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Etag").
						Checks(func(v string) bool {
							etagDir = strings.Trim(v, `"`)
							return v != ""
						}),
				),
		},
	}
	RunWithSpecs(t, testsA, specs.PathGatewayUnixFS)

	testsB := SugarTests{
		{
			Name: "GET for /ipfs/ dir listing with matching strong Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", `"{{etag}}"`, etagDir),
				),
			Response: Expect().
				Status(304),
		},
		{
			Name: "GET for /ipfs/ dir listing with matching weak Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("/ipfs/{{cid}}/root2/root3/", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", `W/"{{etag}}"`, etagDir),
				),
			Response: Expect().
				Status(304),
		},
	}
	RunWithSpecs(t, testsB, specs.PathGatewayUnixFS)
}

func TestGatewayCacheWithIPNS(t *testing.T) {
	tooling.LogTestGroup(t, GroupIPNS)

	fixture := car.MustOpenUnixfsCar("gateway-cache/fixtures.car")
	ipns := ipns.MustOpenIPNSRecordWithKey("gateway-cache/k51qzi5uqu5dlxdsdu5fpuu7h69wu4ohp32iwm9pdt9nq3y5rpn3ln9j12zfhe.ipns-record")
	ipnsKey := ipns.Key()

	tests := SugarTests{
		{
			Name: "GET for /ipns/ unixfs dir listing succeeds",
			Request: Request().
				Path("/ipns/{{KEY}}/root2/root3/", ipnsKey),
			Response: AllOf(
				Expect().
					Status(200).
					Headers(
						Header("X-Ipfs-Path").
							Equals("/ipns/{{KEY}}/root2/root3/", ipnsKey),
						Header("X-Ipfs-Roots").
							Equals("{{CID1}},{{CID2}},{{CID3}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3")),
						Header("Etag").
							Matches("DirIndex-.*_CID-{{CID}}", fixture.MustGetCid("root2", "root3")),
					),
				AnyOf(
					Expect().Headers(Header("Cache-Control").IsEmpty()),
					Expect().Headers(Header("Cache-Control").Matches("public, max-age=*")),
				),
			),
		},
		{
			Name: "GET for /ipns/ unixfs dir with index.html succeeds",
			Request: Request().
				Path("/ipns/{{KEY}}/root2/root3/root4/", ipnsKey),
			Response: AllOf(
				Expect().
					Status(200).
					Headers(
						Header("X-Ipfs-Path").
							Equals("/ipns/{{KEY}}/root2/root3/root4/", ipnsKey),
						Header("X-Ipfs-Roots").
							Equals("{{CID1}},{{CID2}},{{CID3}},{{CID4}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4")),
						Header("Etag").
							Matches(`"{{CID}}"`, fixture.MustGetCid("root2", "root3", "root4")),
					),
				AnyOf(
					Expect().Headers(Header("Cache-Control").IsEmpty()),
					Expect().Headers(Header("Cache-Control").Matches("public, max-age=*")),
				),
			),
		},
		{
			Name: "GET for /ipns/ unixfs file succeeds",
			Request: Request().
				Path("/ipns/{{KEY}}/root2/root3/root4/index.html", ipnsKey),
			Response: AllOf(
				Expect().
					Status(200).
					Headers(
						Header("X-Ipfs-Path").
							Equals("/ipns/{{KEY}}/root2/root3/root4/index.html", ipnsKey),
						Header("X-Ipfs-Roots").
							Equals("{{CID1}},{{CID2}},{{CID3}},{{CID4}},{{CID5}}", fixture.MustGetCid(), fixture.MustGetCid("root2"), fixture.MustGetCid("root2", "root3"), fixture.MustGetCid("root2", "root3", "root4"), fixture.MustGetCid("root2", "root3", "root4", "index.html")),
						Header("Etag").
							Equals(`"{{CID}}"`, fixture.MustGetCid("root2", "root3", "root4", "index.html")),
					),
				AnyOf(
					Expect().Headers(Header("Cache-Control").IsEmpty()),
					Expect().Headers(Header("Cache-Control").Matches("public, max-age=*")),
				),
			),
		},
		{
			Name: "GET for /ipns/ unixfs dir as raw block succeeds",
			Request: Request().
				Path("/ipns/{{KEY}}/root2/root3/root4/", ipnsKey).
				Query("format", "raw"),
			Response: AllOf(
				Expect().
					Status(200).
					Body(
						IsEqualBytes([]byte{
							0x12, 0x34, 0x0a, 0x24, 0x01, 0x55, 0x12, 0x20, 0x58, 0x91, 0xb5, 0xb5,
							0x22, 0xd5, 0xdf, 0x08, 0x6d, 0x0f, 0xf0, 0xb1, 0x10, 0xfb, 0xd9, 0xd2,
							0x1b, 0xb4, 0xfc, 0x71, 0x63, 0xaf, 0x34, 0xd0, 0x82, 0x86, 0xa2, 0xe8,
							0x46, 0xf6, 0xbe, 0x03, 0x12, 0x0a, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2e,
							0x68, 0x74, 0x6d, 0x6c, 0x18, 0x06, 0x0a, 0x02, 0x08, 0x01,
						}),
					),
				AnyOf(
					Expect().Headers(Header("Cache-Control").IsEmpty()),
					Expect().Headers(Header("Cache-Control").Matches("public, max-age=*")),
				),
			),
		},
		{
			Name: "GET for /ipns/ file with matching Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("/ipns/{{KEY}}/root2/root3/root4/index.html", ipnsKey).
				Headers(
					Header("If-None-Match", Fmt(`"{{CID}}"`, fixture.MustGetCid("root2", "root3", "root4", "index.html"))),
				),
			Response: Expect().
				Status(304),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS, specs.PathGatewayIPNS)
}

func TestGatewaySymlink(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("path_gateway_unixfs/symlink.car")
	rootDirCID := fixture.MustGetCid()

	tests := SugarTests{
		{
			Name: "Test the directory listing",
			Request: Request().
				Path("/ipfs/{{CID}}/", rootDirCID),
			Response: Expect().
				Body(
					And(
						Contains(">foo<"),
						Contains(">bar<"),
					),
				),
		},
		{
			Name: "Test the directory raw query",
			Request: Request().
				Path("/ipfs/{{CID}}", rootDirCID).
				Query("format", "raw"),
			Response: Expect().
				Status(200).
				Body(fixture.MustGetRawData()),
		},
		{
			Name: "Test the symlink",
			Request: Request().
				Path("/ipfs/{{CID}}/bar", rootDirCID),
			Response: Expect().
				Status(200).
				Bytes("foo"),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS)
}

func TestGatewayUnixFSFileRanges(t *testing.T) {
	tooling.LogTestGroup(t, GroupUnixFS)

	// Range requests MUST conform to the HTTP semantics. The server does not
	// need to be able to support returning multiple ranges. However, it must respond
	// correctly.
	fixture := car.MustOpenUnixfsCar("path_gateway_unixfs/dir-with-files.car")

	var (
		contentType  string
		contentRange string
	)

	RunWithSpecs(t, SugarTests{
		{
			Name: "GET for /ipfs/ file includes Accept-Ranges header",
			Hint: "Gateway returns explicit hint that range requests are supported. This is important for interop with HTTP reverse proxies, CDNs, caches.",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#accept-ranges-response-header",
			Request: Request().
				Path("/ipfs/{{cid}}/ascii.txt", fixture.MustGetCid()),
			Response: Expect().
				Status(200).
				Headers(
					Header("Accept-Ranges").Equals("bytes"),
				).
				Body(fixture.MustGetRawData("ascii.txt")),
		},
		{
			Name: "GET for /ipfs/ file with single range request includes correct bytes",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#range-request-header",
			Request: Request().
				Path("/ipfs/{{cid}}/ascii.txt", fixture.MustGetCid()).
				Headers(
					Header("Range", "bytes=6-16"),
				),
			Response: Expect().
				Status(206).
				Headers(
					Header("Content-Range").Equals("bytes 6-16/31"),
				).
				Body(fixture.MustGetRawData("ascii.txt")[6:17]),
		},
		{
			Name: "GET for /ipfs/ file with suffix range request includes correct bytes from the end of file",
			Hint: "Ensures it is possible to read the tail of a file",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#range-request-header",
			Request: Request().
				Path("/ipfs/{{cid}}/ascii.txt", fixture.MustGetCid()).
				Headers(
					Header("Range", "bytes=-3"),
				),
			Response: Expect().
				Status(206).
				Headers(
					Header("Content-Range").Equals("bytes 28-30/31"),
				).
				Body(fixture.MustGetRawData("ascii.txt")[28:31]),
		},
		{
			Name: "GET for /ipfs/ file with multiple range request returned HTTP 206",
			Hint: "This test reads Content-Type and Content-Range of response, which enable later tests to check if response was acceptable (either single range, or multiple ones)",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#range-request-header",
			Request: Request().
				Path("/ipfs/{{cid}}/ascii.txt", fixture.MustGetCid()).
				Headers(
					Header("Range", "bytes=6-16,0-4"),
				),
			Response: Expect().
				Status(206).
				Headers(
					Header("Content-Type").
						Checks(func(v string) bool {
							// Not really a test, just inspect value
							contentType = v
							return true
						}),
					Header("Content-Range").
						ChecksAll(func(v []string) bool {
							// Not really a test, just inspect value
							if len(v) == 1 {
								contentRange = v[0]

							}
							return true
						}),
				),
		},
	}, specs.PathGatewayRaw)

	tests := SugarTests{}

	multipleRangeSupported := strings.Contains(contentType, "multipart/byteranges")
	onlySingleRangeSupported := !multipleRangeSupported && contentRange != ""

	if onlySingleRangeSupported {
		// Server supports range requests but only the first range.
		tests = append(tests, SugarTest{
			Name: "GET for /ipfs/ file with multiple range request returns correct bytes for the first range",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#range-request-header",
			Request: Request().
				Path("/ipfs/{{cid}}/ascii.txt", fixture.MustGetCid()).
				Headers(
					Header("Range", "bytes=6-16,0-4"),
				),
			Response: Expect().
				Status(206).
				Headers(
					Header("Content-Range", "bytes 6-16/31"),
				).
				Body(fixture.MustGetRawData("ascii.txt")[6:17]),
		})
	} else if multipleRangeSupported {
		// The server supports responding with multi-range requests.
		tests = append(tests, SugarTest{
			Name: "GET for /ipfs/ file with multiple range request includes correct bytes",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#range-request-header",
			Request: Request().
				Path("/ipfs/{{cid}}/ascii.txt", fixture.MustGetCid()).
				Headers(
					Header("Range", "bytes=6-16,0-4"),
				),
			Response: Expect().
				Status(206).
				Headers(
					Header("Content-Type").Contains("multipart/byteranges"),
				).
				Body(And(
					Contains("Content-Range: bytes 6-16/31"),
					Contains(string(fixture.MustGetRawData("ascii.txt")[6:17])),
					Contains("Content-Range: bytes 0-4/31"),
					Contains(string(fixture.MustGetRawData("ascii.txt")[0:5])),
				)),
		})
	} else {
		t.Error("Content-Range and Content-Type header did not match any of the accepted options for a Range request (neither single or multiple ranges are supported)")
	}

	// Range request should work when unrelated parts of DAG not available.
	missingBlockFixture := car.MustOpenUnixfsCar("trustless_gateway_car/file-3k-and-3-blocks-missing-block.car")
	tests = append(tests, SugarTest{
		Name: "GET Range of file succeeds even if the gateway is missing a block AFTER the requested range",
		Hint: "This MUST succeed despite the fact that bytes beyond the end of range are not retrievable",
		Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#range-request-header",
		Request: Request().
			Path("/ipfs/{{cid}}", missingBlockFixture.MustGetCidWithCodec(0x70)).
			Headers(
				Header("Range", "bytes=997-1000"),
			),
		Response: Expect().
			Status(206).
			Headers(
				Header("Content-Range").Equals("bytes 997-1000/3072"),
			),
	}, SugarTest{
		Name: "GET Range of file succeeds even if the gateway is missing a block BEFORE the requested range",
		Hint: "This MUST succeed despite the fact that bytes beyond the end of range are not retrievable",
		Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#range-request-header",
		Request: Request().
			Path("/ipfs/{{cid}}", missingBlockFixture.MustGetCidWithCodec(0x70)).
			Headers(
				Header("Range", "bytes=2200-2201"),
			),
		Response: Expect().
			Status(206).
			Headers(
				Header("Content-Range").Equals("bytes 2200-2201/3072"),
			),
	},
	)

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS)
}

func TestPathGatewayMiscellaneous(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("path_gateway_unixfs/dir-with-percent-encoded-filename.car")
	rootDirCID := fixture.MustGetCid()

	tests := SugarTests{
		{
			Name: "GET for /ipfs/ file whose filename contains percentage-encoded characters works",
			Request: Request().
				Path("/ipfs/{{CID}}/Portugal%252C+España=Peninsula%20Ibérica.txt", rootDirCID),
			Response: Expect().
				Body(fixture.MustGetRawData("Portugal%2C+España=Peninsula Ibérica.txt")),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS)
}
