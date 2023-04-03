package tests

import (
	"fmt"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

/* TODO
1.
test_expect_success "GET for /ipns/ unixfs dir listing succeeds" '
curl -svX GET "http://127.0.0.1:$GWAY_PORT/ipns/$TEST_IPNS_ID/root2/root3/" >/dev/null 2>curl_ipns_dir_listing_output
'
test_expect_success "GET /ipns/ unixfs dir listing has no Cache-Control" '
test_should_not_contain "< Cache-Control" curl_ipns_dir_listing_output
'
test_expect_success "GET /ipns/ unixfs dir listing has no Cache-Control" '
test_should_not_contain "< Cache-Control" curl_ipns_dir_listing_output
'
test_expect_success "GET /ipns/ dir listing response has original content path in X-Ipfs-Path" '
test_should_contain "< X-Ipfs-Path: /ipns/$TEST_IPNS_ID/root2/root3" curl_ipns_dir_listing_output
'
test_expect_success "GET /ipns/ dir listing response has logical CID roots in X-Ipfs-Roots" '
test_should_contain "< X-Ipfs-Roots: ${ROOT1_CID},${ROOT2_CID},${ROOT3_CID}" curl_ipns_dir_listing_output
'
test_expect_success "GET /ipns/ dir response has special Etag for generated dir listing" '
test_should_contain "< Etag: \"DirIndex" curl_ipns_dir_listing_output &&
grep -E "< Etag: \"DirIndex-.+_CID-${ROOT3_CID}\"" curl_ipns_dir_listing_output
'
2.
test_expect_success "GET for /ipns/ unixfs dir with index.html succeeds" '
curl -svX GET "http://127.0.0.1:$GWAY_PORT/ipns/$TEST_IPNS_ID/root2/root3/root4/" >/dev/null 2>curl_ipns_dir_index.html_output
'
test_expect_success "GET /ipns/ unixfs dir with index.html has no Cache-Control" '
test_should_not_contain "< Cache-Control" curl_ipns_dir_index.html_output
'
test_expect_success "GET /ipns/ dir index.html response has original content path in X-Ipfs-Path" '
test_should_contain "< X-Ipfs-Path: /ipns/$TEST_IPNS_ID/root2/root3/root4/" curl_ipns_dir_index.html_output
'
test_expect_success "GET /ipns/ dir index.html response has logical CID roots in X-Ipfs-Roots" '
test_should_contain "< X-Ipfs-Roots: ${ROOT1_CID},${ROOT2_CID},${ROOT3_CID},${ROOT4_CID}" curl_ipns_dir_index.html_output
'
test_expect_success "GET /ipns/ dir index.html response has dir CID as Etag" '
test_should_contain "< Etag: \"${ROOT4_CID}\"" curl_ipns_dir_index.html_output
'
3.
test_expect_success "GET for /ipns/ unixfs file succeeds" '
curl -svX GET "http://127.0.0.1:$GWAY_PORT/ipns/$TEST_IPNS_ID/root2/root3/root4/index.html" >/dev/null 2>curl_ipns_file_output
'
test_expect_success "GET /ipns/ unixfs file has no Cache-Control" '
test_should_not_contain "< Cache-Control" curl_ipns_file_output
'
test_expect_success "GET /ipns/ file response has original content path in X-Ipfs-Path" '
test_should_contain "< X-Ipfs-Path: /ipns/$TEST_IPNS_ID/root2/root3/root4/index.html" curl_ipns_file_output
'
test_expect_success "GET /ipns/ file response has logical CID roots in X-Ipfs-Roots" '
test_should_contain "< X-Ipfs-Roots: ${ROOT1_CID},${ROOT2_CID},${ROOT3_CID},${ROOT4_CID},${FILE_CID}" curl_ipns_file_output
'
test_expect_success "GET /ipns/ response has CID as Etag for a file" '
test_should_contain "< Etag: \"${FILE_CID}\"" curl_ipns_file_output
'
4.
test_expect_success "GET for /ipns/ unixfs dir as DAG-JSON succeeds" '
curl -svX GET "http://127.0.0.1:$GWAY_PORT/ipns/$TEST_IPNS_ID/root2/root3/root4/?format=dag-json" >/dev/null 2>curl_ipns_dir_dag-json_output
'
test_expect_success "GET /ipns/ unixfs dir as dag-json has no Cache-Control" '
test_should_not_contain "< Cache-Control" curl_ipns_dir_dag-json_output
'
5.
test_expect_success "GET for /ipns/ unixfs dir as DAG-JSON succeeds" '
curl -svX GET "http://127.0.0.1:$GWAY_PORT/ipns/$TEST_IPNS_ID/root2/root3/root4/?format=json" >/dev/null 2>curl_ipns_dir_json_output
'
test_expect_success "GET /ipns/ unixfs dir as json has no Cache-Control" '
test_should_not_contain "< Cache-Control" curl_ipns_dir_json_output
'
6.
test_expect_success "GET for /ipns/ file with matching Etag in If-None-Match returns 304 Not Modified" '
curl -svX GET -H "If-None-Match: \"$FILE_CID\"" "http://127.0.0.1:$GWAY_PORT/ipns/$TEST_IPNS_ID/root2/root3/root4/index.html" >/dev/null 2>curl_output &&
test_should_contain "304 Not Modified" curl_output
'
7.
# DirIndex etag is based on xxhash(./assets/dir-index-html), so we need to fetch it dynamically
test_expect_success "GET for /ipfs/ dir listing with matching strong Etag in If-None-Match returns 304 Not Modified" '
curl -Is "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/"| grep -i Etag | cut -f2- -d: | tr -d "[:space:]\"" > dir_index_etag &&
curl -svX GET -H "If-None-Match: \"$(<dir_index_etag)\"" "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/" >/dev/null 2>curl_output &&
test_should_contain "304 Not Modified" curl_output
'
8.
test_expect_success "GET for /ipfs/ dir listing with matching weak Etag in If-None-Match returns 304 Not Modified" '
curl -Is "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/"| grep -i Etag | cut -f2- -d: | tr -d "[:space:]\"" > dir_index_etag &&
curl -svX GET -H "If-None-Match: W/\"$(<dir_index_etag)\"" "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/" >/dev/null 2>curl_output &&
test_should_contain "304 Not Modified" curl_output
'
*/
func TestGatewayCache(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0116-gateway-cache.car")

	tests := SugarTests{
		/*
		test_expect_success "GET for /ipfs/ unixfs dir listing succeeds" '
    curl -svX GET "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/" >/dev/null 2>curl_ipfs_dir_listing_output
    '
		test_expect_success "GET /ipfs/ unixfs dir listing has no Cache-Control" '
    test_should_not_contain "< Cache-Control" curl_ipfs_dir_listing_output
    '
		test_expect_success "GET /ipfs/ dir listing response has original content path in X-Ipfs-Path" '
    test_should_contain "< X-Ipfs-Path: /ipfs/$ROOT1_CID/root2/root3" curl_ipfs_dir_listing_output
    '
		test_expect_success "GET /ipfs/ dir listing response has logical CID roots in X-Ipfs-Roots" '
    test_should_contain "< X-Ipfs-Roots: ${ROOT1_CID},${ROOT2_CID},${ROOT3_CID}" curl_ipfs_dir_listing_output
    '
		test_expect_success "GET /ipfs/ dir response has special Etag for generated dir listing" '
    test_should_contain "< Etag: \"DirIndex" curl_ipfs_dir_listing_output &&
    grep -E "< Etag: \"DirIndex-.+_CID-${ROOT3_CID}\"" curl_ipfs_dir_listing_output
    '
		*/
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
		/*
		test_expect_success "GET for /ipfs/ unixfs dir with index.html succeeds" '
    curl -svX GET "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/root4/" >/dev/null 2>curl_ipfs_dir_index.html_output
    '
    test_expect_success "GET /ipfs/ unixfs dir with index.html has expected Cache-Control" '
    test_should_contain "< Cache-Control: public, max-age=29030400, immutable" curl_ipfs_dir_index.html_output
    '
		test_expect_success "GET /ipfs/ dir index.html response has original content path in X-Ipfs-Path" '
    test_should_contain "< X-Ipfs-Path: /ipfs/$ROOT1_CID/root2/root3/root4/" curl_ipfs_dir_index.html_output
    '
		test_expect_success "GET /ipfs/ dir index.html response has logical CID roots in X-Ipfs-Roots" '
    test_should_contain "< X-Ipfs-Roots: ${ROOT1_CID},${ROOT2_CID},${ROOT3_CID},${ROOT4_CID}" curl_ipfs_dir_index.html_output
    '
		test_expect_success "GET /ipfs/ dir index.html response has dir CID as Etag" '
    test_should_contain "< Etag: \"${ROOT4_CID}\"" curl_ipfs_dir_index.html_output
    '
		*/
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
		/*
		test_expect_success "GET for /ipfs/ unixfs file succeeds" '
    curl -svX GET "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/root4/index.html" >/dev/null 2>curl_ipfs_file_output
    '
		test_expect_success "GET /ipfs/ unixfs file has expected Cache-Control" '
    test_should_contain "< Cache-Control: public, max-age=29030400, immutable" curl_ipfs_file_output
    '
		test_expect_success "GET /ipfs/ file response has original content path in X-Ipfs-Path" '
    test_should_contain "< X-Ipfs-Path: /ipfs/$ROOT1_CID/root2/root3/root4/index.html" curl_ipfs_file_output
    '
		test_expect_success "GET /ipfs/ file response has logical CID roots in X-Ipfs-Roots" '
    test_should_contain "< X-Ipfs-Roots: ${ROOT1_CID},${ROOT2_CID},${ROOT3_CID},${ROOT4_CID},${FILE_CID}" curl_ipfs_file_output
    '
		test_expect_success "GET /ipfs/ response has CID as Etag for a file" '
    test_should_contain "< Etag: \"${FILE_CID}\"" curl_ipfs_file_output
    '
		*/
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
		/*
		test_expect_success "GET for /ipfs/ unixfs dir as DAG-JSON succeeds" '
    curl -svX GET "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/root4/?format=dag-json" >/dev/null 2>curl_ipfs_dir_dag-json_output
    '
    test_expect_success "GET /ipfs/ dag-json has expected Cache-Control" '
    test_should_contain "< Cache-Control: public, max-age=29030400, immutable" curl_ipfs_dir_dag-json_output
    '
		*/
		{
			Name: "GET for /ipfs/ unixfs dir as DAG-JSON succeeds",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/", fixture.MustGetCid()).
				Query("format", "dag-json"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						Equals("public, max-age=29030400, immutable"),
				),
		},
		/*
		test_expect_success "GET for /ipfs/ unixfs dir as DAG-JSON succeeds" '
    curl -svX GET "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/root4/?format=json" >/dev/null 2>curl_ipfs_dir_json_output
    '
    test_expect_success "GET /ipfs/ unixfs dir as json has expected Cache-Control" '
    test_should_contain "< Cache-Control: public, max-age=29030400, immutable" curl_ipfs_dir_json_output
    '
		*/
		{
			Name: "GET for /ipfs/ unixfs dir as JSON succeeds",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/", fixture.MustGetCid()).
				Query("format", "json"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Cache-Control").
						Equals("public, max-age=29030400, immutable"),
				),
		},
		/*
		test_expect_success "HEAD for /ipfs/ with only-if-cached succeeds when in local datastore" '
    curl -sv -I -H "Cache-Control: only-if-cached" "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/root4/index.html" > curl_onlyifcached_postitive_head 2>&1 &&
    cat curl_onlyifcached_postitive_head &&
    grep "< HTTP/1.1 200 OK" curl_onlyifcached_postitive_head
    '
		*/
		{
			Name: "HEAD for /ipfs/ with only-if-cached succeeds when in local datastore",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/?format=json", fixture.MustGetCid()).
				Headers(
					Header("Cache-Control", "only-if-cached"),
				).
				Method("HEAD"),
			Response: Expect().
				Status(200),
		},
		/*
		test_expect_success "HEAD for /ipfs/ with only-if-cached fails when not in local datastore" '
    curl -sv -I -H "Cache-Control: only-if-cached" "http://127.0.0.1:$GWAY_PORT/ipfs/$(date | ipfs add --only-hash -Q)" > curl_onlyifcached_negative_head 2>&1 &&
    cat curl_onlyifcached_negative_head &&
    grep "< HTTP/1.1 412 Precondition Failed" curl_onlyifcached_negative_head
    '
		*/
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
		/*
		test_expect_success "GET for /ipfs/ with only-if-cached succeeds when in local datastore" '
    curl -svX GET -H "Cache-Control: only-if-cached" "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/root4/index.html" >/dev/null 2>curl_onlyifcached_postitive_out &&
    cat curl_onlyifcached_postitive_out &&
    grep "< HTTP/1.1 200 OK" curl_onlyifcached_postitive_out
    '
		*/
		{
			Name: "GET for /ipfs/ with only-if-cached succeeds when in local datastore",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/?format=json", fixture.MustGetCid()).
				Headers(
					Header("Cache-Control", "only-if-cached"),
				),
			Response: Expect().
				Status(200),
		},
		/*
		test_expect_success "GET for /ipfs/ with only-if-cached fails when not in local datastore" '
    curl -svX GET -H "Cache-Control: only-if-cached" "http://127.0.0.1:$GWAY_PORT/ipfs/$(date | ipfs add --only-hash -Q)" >/dev/null 2>curl_onlyifcached_negative_out &&
    cat curl_onlyifcached_negative_out &&
    grep "< HTTP/1.1 412 Precondition Failed" curl_onlyifcached_negative_out
    '
		*/
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
		/*
		test_expect_success "GET for /ipfs/ file with matching Etag in If-None-Match returns 304 Not Modified" '
    curl -svX GET -H "If-None-Match: \"$FILE_CID\"" "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/root4/index.html" >/dev/null 2>curl_output &&
    test_should_contain "304 Not Modified" curl_output
    '
		*/
		{
			Name: "GET for /ipfs/ file with matching Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", fmt.Sprintf("\"%s\"", fixture.MustGetCid("root2", "root3", "root4", "index.html"))),
				),
			Response: Expect().
				Status(304),
		},
		/*
		test_expect_success "GET for /ipfs/ dir with index.html file with matching Etag in If-None-Match returns 304 Not Modified" '
    curl -svX GET -H "If-None-Match: \"$ROOT4_CID\"" "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/root4/" >/dev/null 2>curl_output &&
    test_should_contain "304 Not Modified" curl_output
    '
		*/
		{
			Name: "GET for /ipfs/ dir with index.html file with matching Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", fmt.Sprintf("\"%s\"", fixture.MustGetCid("root2", "root3", "root4"))),
				),
			Response: Expect().
				Status(304),
		},
		/*
		test_expect_success "GET for /ipfs/ file with matching third Etag in If-None-Match returns 304 Not Modified" '
    curl -svX GET -H "If-None-Match: \"fakeEtag1\", \"fakeEtag2\", \"$FILE_CID\"" "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/root4/index.html" >/dev/null 2>curl_output &&
    test_should_contain "304 Not Modified" curl_output
    '
		*/
		{
			Name: "GET for /ipfs/ file with matching third Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", fmt.Sprintf("\"fakeEtag1\", \"fakeEtag2\", \"%s\"", fixture.MustGetCid("root2", "root3", "root4", "index.html"))),
				),
			Response: Expect().
				Status(304),
		},
		/*
		test_expect_success "GET for /ipfs/ file with matching weak Etag in If-None-Match returns 304 Not Modified" '
    curl -svX GET -H "If-None-Match: W/\"$FILE_CID\"" "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/root4/index.html" >/dev/null 2>curl_output &&
    test_should_contain "304 Not Modified" curl_output
    '
		*/
		{
			Name: "GET for /ipfs/ file with matching weak Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", fmt.Sprintf("W/\"%s\"", fixture.MustGetCid("root2", "root3", "root4", "index.html"))),
				),
			Response: Expect().
				Status(304),
		},
		/*
		test_expect_success "GET for /ipfs/ file with wildcard Etag in If-None-Match returns 304 Not Modified" '
    curl -svX GET -H "If-None-Match: *" "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/root4/index.html" >/dev/null 2>curl_output &&
    test_should_contain "304 Not Modified" curl_output
    '
		*/
		{
			Name: "GET for /ipfs/ file with wildcard Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/%s/root2/root3/root4/index.html", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", "*"),
				),
			Response: Expect().
				Status(304),
		},
		/*
		test_expect_success "GET for /ipfs/ dir listing with matching weak Etag in If-None-Match returns 304 Not Modified" '
    curl -svX GET -H "If-None-Match: W/\"$ROOT3_CID\"" "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT1_CID/root2/root3/" >/dev/null 2>curl_output &&
    test_should_contain "304 Not Modified" curl_output
    '
		*/
		{
			Name: "GET for /ipfs/ dir listing with matching weak Etag in If-None-Match returns 304 Not Modified",
			Request: Request().
				Path("ipfs/%s/root2/root3/", fixture.MustGetCid()).
				Headers(
					Header("If-None-Match", fmt.Sprintf("W/\"%s\"", fixture.MustGetCid("root2", "root3"))),
				),
			Response: Expect().
				Status(304),
		},
	}.Build()

	Run(t, tests)
}
