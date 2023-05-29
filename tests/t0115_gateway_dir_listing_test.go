package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestDirectorListingOnGateway(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0115/fixtures.car")
	root := fixture.MustGetNode()
	file := fixture.MustGetNode("ą", "ę", "file-źł.txt")

	tests := SugarTests{
		// ## ============================================================================
		// ## Test dir listing on path gateway (eg. 127.0.0.1:8080/ipfs/)
		// ## ============================================================================
		// test_expect_success "path gw: backlink on root CID should be hidden" '
		//
		//	curl -sD - http://127.0.0.1:$GWAY_PORT/ipfs/${DIR_CID}/ > list_response &&
		//	test_should_contain "Index of" list_response &&
		//	test_should_not_contain "<a href=\"/ipfs/$DIR_CID/\">..</a>" list_response,
		//
		// '
		{
			Name: "path gw: backlink on root CID should be hidden",
			Request: Request().
				Path("ipfs/{{cid}}", root.Cid()),
			Response: Expect().
				Body(
					And(
						Contains("Index of"),
						Not(Contains(`<a href="/ipfs/{{cid}}/">..</a>`, root.Cid())),
					)),
		},
		// test_expect_success "path gw: redirect dir listing to URL with trailing slash" '
		//   curl -sD - http://127.0.0.1:$GWAY_PORT/ipfs/${DIR_CID}/ą/ę > list_response &&
		//   test_should_contain "HTTP/1.1 301 Moved Permanently" list_response &&
		//   test_should_contain "Location: /ipfs/${DIR_CID}/%c4%85/%c4%99/" list_response
		// '
		{
			Name: "path gw: redirect dir listing to URL with trailing slash WHAT",
			Request: Request().
				DoNotFollowRedirects().
				Path("ipfs/{{cid}}/ą/ę", root.Cid()),
			Response: Expect().
				Status(301).
				Headers(
					Header("Location", `/ipfs/{{cid}}/%c4%85/%c4%99/`, root.Cid()),
				),
		},
		// test_expect_success "path gw: Etag should be present" '
		//   curl -sD - http://127.0.0.1:$GWAY_PORT/ipfs/${DIR_CID}/ą/ę/ > list_response &&
		//   test_should_contain "Index of" list_response &&
		//   test_should_contain "Etag: \"DirIndex-" list_response
		// '
		// test_expect_success "path gw: breadcrumbs should point at /ipfs namespace mounted at Origin root" '
		//   test_should_contain "/ipfs/<a href=\"/ipfs/$DIR_CID\">$DIR_CID</a>/<a href=\"/ipfs/$DIR_CID/%C4%85\">ą</a>/<a href=\"/ipfs/$DIR_CID/%C4%85/%C4%99\">ę</a>" list_response
		// '
		// test_expect_success "path gw: backlink on subdirectory should point at parent directory" '
		//   test_should_contain "<a href=\"/ipfs/$DIR_CID/%C4%85/%C4%99/..\">..</a>" list_response
		// '
		// test_expect_success "path gw: name column should be a link to its content path" '
		//   test_should_contain "<a href=\"/ipfs/$DIR_CID/%C4%85/%C4%99/file-%C5%BA%C5%82.txt\">file-źł.txt</a>" list_response
		// '
		// test_expect_success "path gw: hash column should be a CID link with filename param" '
		//   test_should_contain "<a class=\"ipfs-hash\" translate=\"no\" href=\"/ipfs/$FILE_CID?filename=file-%25C5%25BA%25C5%2582.txt\">" list_response
		// '
		{
			Name: "path gw: dir listing",
			Request: Request().
				DoNotFollowRedirects().
				Path("ipfs/{{cid}}/ą/ę/", root.Cid()),
			Response: Expect().
				Headers(
					Header("Etag").Contains(`"DirIndex-`),
				).
				BodyWithHint(`
				- Breadcrumbs should point at /ipfs namespace mounted at Origin root
				- backlink on subdirectory should point at parent directory
				- name column should be a link to its content path
				- hash column should be a CID link with filename param
				`,
					And(Contains(`/ipfs/<a href="/ipfs/{{cid}}">{{cid}}</a>/<a href="/ipfs/{{cid}}/%C4%85">ą</a>/<a href="/ipfs/{{cid}}/%C4%85/%C4%99">ę</a>`,
						root.Cid()),
						Contains(`<a href="/ipfs/{{cid}}/%C4%85/%C4%99/..">..</a>`, root.Cid()),
						Contains(`<a href="/ipfs/{{cid}}/%C4%85/%C4%99/file-%C5%BA%C5%82.txt">file-źł.txt</a>`, root.Cid()),
						Contains(`<a class="ipfs-hash" translate="no" href="/ipfs/{{cid}}?filename=file-%25C5%25BA%25C5%2582.txt">`, file.Cid())),
				),
		},
	}

	RunIfSpecsAreEnabled(
		t,
		tests,
	)
}
