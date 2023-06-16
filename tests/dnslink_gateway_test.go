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

// TODO(laurent): this were in t0114_gateway_subdomains_test.go

func TestDNSLinkGateway(t *testing.T) {
	tests := SugarTests{
		// ## ============================================================================
		// ## Test DNSLink requests with a custom PublicGateway (hostname config)
		// ## (DNSLink site at http://dnslink-test.example.com)
		// ## ============================================================================
		// # disable wildcard DNSLink gateway
		// # and enable it on specific NSLink hostname
		// ipfs config --json Gateway.NoDNSLink true && \
		// ipfs config --json Gateway.PublicGateways '{
		//   "dnslink-enabled-on-fqdn.example.org": {
		//     "NoDNSLink": false,
		//     "UseSubdomains": false,
		//     "Paths": ["/ipfs"]
		//   },
		//   "only-dnslink-enabled-on-fqdn.example.org": {
		//     "NoDNSLink": false,
		//     "UseSubdomains": false,
		//     "Paths": []
		//   },
		//   "dnslink-disabled-on-fqdn.example.com": {
		//     "NoDNSLink": true,
		//     "UseSubdomains": false,
		//     "Paths": []
		//   }
		// }' || exit 1

		// # DNSLink test requires a daemon in online mode with precached /ipns/ mapping
		// DNSLINK_FQDN="dnslink-enabled-on-fqdn.example.org"
		// ONLY_DNSLINK_FQDN="only-dnslink-enabled-on-fqdn.example.org"
		// NO_DNSLINK_FQDN="dnslink-disabled-on-fqdn.example.com"
		// export IPFS_NS_MAP="$DNSLINK_FQDN:/ipfs/$CIDv1,$ONLY_DNSLINK_FQDN:/ipfs/$DIR_CID"

		// # DNSLink enabled

		// test_hostname_gateway_response_should_contain \
		//   "request for http://{dnslink-fqdn}/ PublicGateway returns expected payload" \
		//   "$DNSLINK_FQDN" \
		//   "http://127.0.0.1:$GWAY_PORT/" \
		//   "$CID_VAL"

		// test_hostname_gateway_response_should_contain \
		//   "request for {dnslink-fqdn}/ipfs/{cid} returns expected payload when /ipfs is on Paths whitelist" \
		//   "$DNSLINK_FQDN" \
		//   "http://127.0.0.1:$GWAY_PORT/ipfs/$CIDv1" \
		//   "$CID_VAL"

		// # Test for a fun edge case: DNSLink-only gateway without  /ipfs/ namespace
		// # mounted, and with subdirectory named "ipfs" ¯\_(ツ)_/¯
		// test_hostname_gateway_response_should_contain \
		//   "request for {dnslink-fqdn}/ipfs/file.txt returns data from content root when /ipfs in not on Paths whitelist" \
		//   "$ONLY_DNSLINK_FQDN" \
		//   "http://127.0.0.1:$GWAY_PORT/ipfs/file.txt" \
		//   "I am a txt file"

		// test_hostname_gateway_response_should_contain \
		//   "request for {dnslink-fqdn}/ipns/{peerid} returns 404 when path is not whitelisted" \
		//   "$DNSLINK_FQDN" \
		//   "http://127.0.0.1:$GWAY_PORT/ipns/$RSA_IPNS_IDv0" \
		//   "404 Not Found"

		// test_hostname_gateway_response_should_contain \
		//   "request for {dnslink-fqdn}/ipns/{peerid} returns 404 when path is not whitelisted" \
		//   "$DNSLINK_FQDN" \
		//   "http://127.0.0.1:$GWAY_PORT/ipns/$ED25519_IPNS_IDv0" \
		//   "404 Not Found"

		// # DNSLink disabled

		// test_hostname_gateway_response_should_contain \
		//   "request for http://{dnslink-fqdn}/ returns 404 when NoDNSLink=true" \
		//   "$NO_DNSLINK_FQDN" \
		//   "http://127.0.0.1:$GWAY_PORT/" \
		//   "404 Not Found"

		// test_hostname_gateway_response_should_contain \
		//   "request for {dnslink-fqdn}/ipfs/{cid} returns 404 when path is not whitelisted" \
		//   "$NO_DNSLINK_FQDN" \
		//   "http://127.0.0.1:$GWAY_PORT/ipfs/$CIDv0" \
		//   "404 Not Found"

		// ## ============================================================================
		// ## Test wildcard DNSLink (any hostname, with default config)
		// ## ============================================================================

		// test_kill_ipfs_daemon

		// # enable wildcard DNSLink gateway (any value in Host header)
		// # and remove custom PublicGateways
		// ipfs config --json Gateway.NoDNSLink false && \
		// ipfs config --json Gateway.PublicGateways '{}' || exit 1

		// # DNSLink test requires a daemon in online mode with precached /ipns/ mapping
		// DNSLINK_FQDN="wildcard-dnslink-not-in-config.example.com"
		// export IPFS_NS_MAP="$DNSLINK_FQDN:/ipfs/$CIDv1"

		// # restart daemon to apply config changes
		// test_launch_ipfs_daemon

		// # make sure test setup is valid (fail if CoreAPI is unable to resolve)
		// test_expect_success "spoofed DNSLink record resolves in cli" "
		//   ipfs resolve /ipns/$DNSLINK_FQDN > result &&
		//   test_should_contain \"$CIDv1\" result &&
		//   ipfs cat /ipns/$DNSLINK_FQDN > result &&
		//   test_should_contain \"$CID_VAL\" result
		// "

		// # gateway test
		//
		//	test_hostname_gateway_response_should_contain \
		//	  "request for http://{dnslink-fqdn}/ (wildcard) returns expected payload" \
		//	  "$DNSLINK_FQDN" \
		//	  "http://127.0.0.1:$GWAY_PORT/" \
		//	  "$CID_VAL"
	}

	RunWithSpecs(t, helpers.UnwrapSubdomainTests(t, tests), specs.DNSLinkGateway)
}

// TODO(laurent): this was in t0115_gateway_dir_listing_test.go

func TestDNSLinkGatewayUnixFSDirectoryListing(t *testing.T) {
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

	RunWithSpecs(t, helpers.UnwrapSubdomainTests(t, tests), specs.DNSLinkGateway)
}
