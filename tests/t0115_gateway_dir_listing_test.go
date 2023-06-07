package tests

import (
	"net/url"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/dnslink"
	"github.com/ipfs/gateway-conformance/tooling/helpers"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
	"github.com/ipfs/gateway-conformance/tooling/tmpl"
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
				Path("ipfs/{{cid}}/ą/ę/", root.Cid()),
			Response: Expect().
				Headers(
					Header("Etag").Contains(`"DirIndex-`),
				).
				BodyWithHint(`
				- should contain "Index of"
				- Breadcrumbs should point at /ipfs namespace mounted at Origin root
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
	}

	RunIfSpecsAreEnabled(
		t,
		tests,
	)
}

func TestDirListingOnSubdomainGateway(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0115/fixtures.car")
	root := fixture.MustGetNode()
	file := fixture.MustGetNode("ą", "ę", "file-źł.txt")

	// We're going to run the same test against multiple gateways (localhost, and a subdomain gateway)
	gatewayURLs := []string{
		SubdomainGatewayURL,
		SubdomainLocalhostGatewayURL,
	}

	tests := SugarTests{}

	for _, gatewayURL := range gatewayURLs {
		u, err := url.Parse(gatewayURL)
		if err != nil {
			t.Fatal(err)
		}

		// ## ============================================================================
		// ## Test dir listing on subdomain gateway (eg. <cid>.ipfs.localhost:8080)
		// ## ============================================================================
		tests = append(tests, SugarTests{
			// DIR_HOSTNAME="${DIR_CID}.ipfs.localhost"
			// # note: we skip DNS lookup by running curl with --resolve $DIR_HOSTNAME:127.0.0.1

			// test_expect_success "subdomain gw: backlink on root CID should be hidden" '
			//   curl -sD - --resolve $DIR_HOSTNAME:$GWAY_PORT:127.0.0.1 http://$DIR_HOSTNAME:$GWAY_PORT/ > list_response &&
			//   test_should_contain "Index of" list_response &&
			//   test_should_not_contain "<a href=\"/\">..</a>" list_response
			// '
			{
				Name: "backlink on root CID should be hidden",
				Request: Request().
					URL(
						"{{scheme}}://{{cid}}.ipfs.{{host}}/",
						u.Scheme,
						root.Cid(),
						u.Host,
					),
				Response: Expect().
					BodyWithHint("backlink on root CID should be hidden",
						And(
							Contains("Index of"),
							Not(Contains(`<a href="/">..</a>`)),
						)),
			},
			// test_expect_success "subdomain gw: redirect dir listing to URL with trailing slash" '
			//   curl -sD - --resolve $DIR_HOSTNAME:$GWAY_PORT:127.0.0.1 http://$DIR_HOSTNAME:$GWAY_PORT/ą/ę > list_response &&
			//   test_should_contain "HTTP/1.1 301 Moved Permanently" list_response &&
			//   test_should_contain "Location: /%c4%85/%c4%99/" list_response
			// '
			{
				Name: "redirect dir listing to URL with trailing slash",
				Request: Request().
					URL(
						"{{scheme}}://{{cid}}.ipfs.{{host}}/ą/ę",
						u.Scheme,
						root.Cid(),
						u.Host,
					),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals(`/%c4%85/%c4%99/`),
					),
			},
			// test_expect_success "subdomain gw: Etag should be present" '
			//   curl -sD - --resolve $DIR_HOSTNAME:$GWAY_PORT:127.0.0.1 http://$DIR_HOSTNAME:$GWAY_PORT/ą/ę/ > list_response &&
			//   test_should_contain "Index of" list_response &&
			//   test_should_contain "Etag: \"DirIndex-" list_response
			// '
			// test_expect_success "subdomain gw: backlink on subdirectory should point at parent directory" '
			//   test_should_contain "<a href=\"/%C4%85/%C4%99/..\">..</a>" list_response
			// '
			// test_expect_success "subdomain gw: breadcrumbs should leverage path-based router mounted on the parent domain" '
			//   test_should_contain "/ipfs/<a href=\"//localhost:$GWAY_PORT/ipfs/$DIR_CID\">$DIR_CID</a>/<a href=\"//localhost:$GWAY_PORT/ipfs/$DIR_CID/%C4%85\">ą</a>/<a href=\"//localhost:$GWAY_PORT/ipfs/$DIR_CID/%C4%85/%C4%99\">ę</a>" list_response
			// '
			// test_expect_success "subdomain gw: name column should be a link to content root mounted at subdomain origin" '
			//   test_should_contain "<a href=\"/%C4%85/%C4%99/file-%C5%BA%C5%82.txt\">file-źł.txt</a>" list_response
			// '
			// test_expect_success "subdomain gw: hash column should be a CID link to path router with filename param" '
			//   test_should_contain "<a class=\"ipfs-hash\" translate=\"no\" href=\"//localhost:$GWAY_PORT/ipfs/$FILE_CID?filename=file-%25C5%25BA%25C5%2582.txt\">" list_response
			// '
			{
				Name: "Regular dir listing",
				Request: Request().URL(
					"{{scheme}}://{{cid}}.ipfs.{{host}}/ą/ę/",
					u.Scheme,
					root.Cid(),
					u.Host,
				),
				Response: Expect().
					Headers(
						Header("Etag").Contains(`"DirIndex-`),
					).BodyWithHint(`
					- backlink on subdirectory should point at parent directory
					- breadcrumbs should leverage path-based router mounted on the parent domain
					- name column should be a link to content root mounted at subdomain origin
					`,
					And(
						Contains("Index of"),
						Contains(
							`<a href="/%C4%85/%C4%99/..">..</a>`,
						),
						Contains(
							`/ipfs/<a href="//{{host}}/ipfs/{{cid}}">{{cid}}</a>/<a href="//{{host}}/ipfs/{{cid}}/%C4%85">ą</a>/<a href="//{{host}}/ipfs/{{cid}}/%C4%85/%C4%99">ę</a>`,
							u.Host, // We don't have a subdomain here which prevents issues with normalization and cidv0
							root.Cid(),
						),
						Contains(
							`<a href="/%C4%85/%C4%99/file-%C5%BA%C5%82.txt">file-źł.txt</a>`,
						),
						Contains(
							`<a class="ipfs-hash" translate="no" href="//{{host}}/ipfs/{{cid}}?filename=file-%25C5%25BA%25C5%2582.txt">`,
							u.Host, // We don't have a subdomain here which prevents issues with normalization and cidv0
							file.Cid(),
						),
					)),
			},
		}...)
	}

	// Body expect to find substring '<a class="ipfs-hash" translate="no" href="//example.com/ipfs/bafybeig6ka5mlwkl4subqhaiatalkcleo4jgnr3hqwvpmsqfca27cijp3i?filename=file-%25C5%25BA%25C5%2582.txt">',

	RunIfSpecsAreEnabled(
		t,
		helpers.UnwrapSubdomainTests(
			t,
			tests,
		),
		specs.SubdomainGateway,
	)
}

func TestDirListingOnDNSLinkGateway(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0115/fixtures.car")
	file := fixture.MustGetNode("ą", "ę", "file-źł.txt")

	dnsLinks := dnslink.MustOpenDNSLink("t0115/dnslink.yml")
	dnsLink := dnsLinks.MustGet("website")

	gatewayURL := SubdomainGatewayURL

	tests := SugarTests{}

	u, err := url.Parse(gatewayURL)
	if err != nil {
		t.Fatal(err)
	}

	dnsLinkHostname := tmpl.Fmt("{{dnslink}}.{{host}}", dnsLink, u.Host)

	// ## ============================================================================
	// ## Test dir listing on DNSLink gateway (eg. example.com)
	// ## ============================================================================
	tests = append(tests, SugarTests{
		// # DNSLink test requires a daemon in online mode with precached /ipns/ mapping
		// test_kill_ipfs_daemon
		// DNSLINK_HOSTNAME="website.example.com"
		// export IPFS_NS_MAP="$DNSLINK_HOSTNAME:/ipfs/$DIR_CID"
		// test_launch_ipfs_daemon

		// # Note that:
		// # - this type of gateway is also tested in gateway_test.go#TestIPNSHostnameBacklinks
		// #   (go tests and sharness tests should be kept in sync)
		// # - we skip DNS lookup by running curl with --resolve $DNSLINK_HOSTNAME:127.0.0.1

		// test_expect_success "dnslink gw: backlink on root CID should be hidden" '
		//   curl -v -sD - --resolve $DNSLINK_HOSTNAME:$GWAY_PORT:127.0.0.1 http://$DNSLINK_HOSTNAME:$GWAY_PORT/ > list_response &&
		//   test_should_contain "Index of" list_response &&
		//   test_should_not_contain "<a href=\"/\">..</a>" list_response
		// '
		{
			Name: "Backlink on root CID should be hidden",
			Request: Request().
				URL(`{{scheme}}://{{hostname}}/`, u.Scheme, dnsLinkHostname),
			Response: Expect().
				Body(
					And(
						Contains("Index of"),
						Not(Contains(`<a href="/">..</a>`)),
					),
				),
		},
		// test_expect_success "dnslink gw: redirect dir listing to URL with trailing slash" '
		//   curl -sD - --resolve $DNSLINK_HOSTNAME:$GWAY_PORT:127.0.0.1 http://$DNSLINK_HOSTNAME:$GWAY_PORT/ą/ę > list_response &&
		//   test_should_contain "HTTP/1.1 301 Moved Permanently" list_response &&
		//   test_should_contain "Location: /%c4%85/%c4%99/" list_response
		// '
		{
			Name: "Redirect dir listing to URL with trailing slash",
			Request: Request().
				URL(`{{scheme}}://{{hostname}}/ą/ę`, u.Scheme, dnsLinkHostname),
			Response: Expect().
				Status(301).
				Headers(
					Header("Location").Equals(`/%c4%85/%c4%99/`),
				),
		},
		// test_expect_success "dnslink gw: Etag should be present" '
		//   curl -sD - --resolve $DNSLINK_HOSTNAME:$GWAY_PORT:127.0.0.1 http://$DNSLINK_HOSTNAME:$GWAY_PORT/ą/ę/ > list_response &&
		//   test_should_contain "Index of" list_response &&
		//   test_should_contain "Etag: \"DirIndex-" list_response
		// '
		// test_expect_success "dnslink gw: backlink on subdirectory should point at parent directory" '
		//   test_should_contain "<a href=\"/%C4%85/%C4%99/..\">..</a>" list_response
		// '
		// test_expect_success "dnslink gw: breadcrumbs should point at content root mounted at dnslink origin" '
		//   test_should_contain "/ipns/<a href=\"//$DNSLINK_HOSTNAME:$GWAY_PORT/\">website.example.com</a>/<a href=\"//$DNSLINK_HOSTNAME:$GWAY_PORT/%C4%85\">ą</a>/<a href=\"//$DNSLINK_HOSTNAME:$GWAY_PORT/%C4%85/%C4%99\">ę</a>" list_response
		// '
		// test_expect_success "dnslink gw: name column should be a link to content root mounted at dnslink origin" '
		//   test_should_contain "<a href=\"/%C4%85/%C4%99/file-%C5%BA%C5%82.txt\">file-źł.txt</a>" list_response
		// '
		// # DNSLink websites don't have public gateway mounted by default
		// # See: https://github.com/ipfs/dir-index-html/issues/42
		// test_expect_success "dnslink gw: hash column should be a CID link to cid.ipfs.tech" '
		//   test_should_contain "<a class=\"ipfs-hash\" translate=\"no\" href=\"https://cid.ipfs.tech/#$FILE_CID\" target=\"_blank\" rel=\"noreferrer noopener\">" list_response
		// '
		{
			Name: "Regular dir listing",
			Request: Request().
				URL(`{{scheme}}://{{hostname}}/ą/ę/`, u.Scheme, dnsLinkHostname),
			Response: Expect().
				Headers(
					Header("Etag").Contains(`"DirIndex-`),
				).
				BodyWithHint(`
					- backlink on subdirectory should point at parent directory
					- breadcrumbs should point at content root mounted at dnslink origin
					- name column should be a link to content root mounted at dnslink origin
					- hash column should be a CID link to cid.ipfs.tech
					  DNSLink websites don't have public gateway mounted by default
					  See: https://github.com/ipfs/dir-index-html/issues/42
					`,
					And(
						Contains("Index of"),
						Contains(`<a href="/%C4%85/%C4%99/..">..</a>`),
						Contains(`/ipns/<a href="//{{hostname}}/">{{hostname}}</a>/<a href="//{{hostname}}/%C4%85">ą</a>/<a href="//{{hostname}}/%C4%85/%C4%99">ę</a>`, dnsLinkHostname),
						Contains(`<a href="/%C4%85/%C4%99/file-%C5%BA%C5%82.txt">file-źł.txt</a>`),
						Contains(`<a class="ipfs-hash" translate="no" href="https://cid.ipfs.tech/#{{cid}}" target="_blank" rel="noreferrer noopener">`, file.Cid()),
					),
				),
		},
	}...)

	RunIfSpecsAreEnabled(
		t,
		helpers.UnwrapSubdomainTests(
			t,
			tests,
		),
		specs.SubdomainGateway,
		specs.DNSLinkResolver,
	)
}
