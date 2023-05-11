package tests

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
	"github.com/ipfs/gateway-conformance/tooling/tmpl"
)

func TestGatewaySubdomains(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0114-gateway_subdomains.car")

	CIDVal := string(fixture.MustGetRawData("hello-CIDv1")) // hello
	DirCID := fixture.MustGetCid("testdirlisting")
	CIDv1 := fixture.MustGetCid("hello-CIDv1")
	CIDv0 := fixture.MustGetCid("hello-CIDv0")
	CIDv0to1 := fixture.MustGetCid("hello-CIDv0to1")
	CIDv1_TOO_LONG := fixture.MustGetCid("hello-CIDv1_TOO_LONG")
	CIDWikipedia := "QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco"

	tests := SugarTests{}

	// sugar: readable way to add more tests
	with := func(moreTests SugarTests) {
		tests = append(tests, moreTests...)
	}

	// sugar: nicer looking sprintf call
	URL := func(path string, args ...interface{}) string {
		return tmpl.Fmt(path, args...)
	}

	// We're going to run the same test against multiple gateways (localhost, and a subdomain gateway)
	gatewayURLs := []string{
		SubdomainGatewayURL,
		SubdomainLocalhostGatewayURL,
	}

	for _, gatewayURL := range gatewayURLs {
		u, err := url.Parse(gatewayURL)
		if err != nil {
			t.Fatal(err)
		}

		with(testGatewayWithManyProtocols(t,
			"request for example.com/ipfs/{CIDv1} redirects to subdomain",
			`
			subdomains should not return payload directly,
			but redirect to URL with proper origin isolation
			`,
			URL("{{url}}/ipfs/{{cid}}/", gatewayURL, CIDv1),
			Expect().
				Status(301).
				Headers(
					Header("Location").
						Hint("request for example.com/ipfs/{CIDv1} returns Location HTTP header for subdomain redirect in browsers").
						Contains("{{scheme}}://{{cid}}.ipfs.{{host}}/", u.Scheme, CIDv1, u.Host),
				).
				BodyWithHint(`
					We return body with HTTP 301 so existing cli scripts that use path-based
					gateway do not break (curl doesn't auto-redirect without passing -L; wget
					does not span across hostnames by default)
					Context: https://github.com/ipfs/go-ipfs/issues/6975
				`,
					IsEqual("hello\n"),
				),
		))

		with(testGatewayWithManyProtocols(t,
			"request for example.com/ipfs/{DirCID} redirects to subdomain",
			`
			subdomains should not return payload directly,
			but redirect to URL with proper origin isolation
			`,
			URL("{{url}}/ipfs/{{cid}}/", gatewayURL, DirCID),
			Expect().
				Status(301).
				Headers(
					Header("Location").
						Hint("request for example.com/ipfs/{DirCID} returns Location HTTP header for subdomain redirect in browsers").
						Contains("{{scheme}}://{{cid}}.ipfs.{{host}}/", u.Scheme, DirCID, u.Host),
				),
		))

		with(testGatewayWithManyProtocols(t,
			"request for example.com/ipfs/{CIDv0} redirects to CIDv1 representation in subdomain",
			"",
			URL("{{url}}/ipfs/{{cid}}/", gatewayURL, CIDv0),
			Expect().
				Status(301).
				Headers(
					Header("Location").
						Hint("request for example.com/ipfs/{CIDv0to1} returns Location HTTP header for subdomain redirect in browsers").
						Contains("{{scheme}}://{{cid}}.ipfs.{{host}}/", u.Scheme, CIDv0to1, u.Host),
				),
		))

		// TODO: ipns
		// TODO: dns link test

		// ============================================================================
		// Test subdomain-based requests to a local gateway with default config
		// (origin per content root at http://*.example.com)
		// ============================================================================

		with(testGatewayWithManyProtocols(t,
			"request for {CID}.ipfs.example.com should return expected payload",
			"",
			URL("{{scheme}}://{{cid}}.ipfs.{{host}}", u.Scheme, CIDv1, u.Host),
			Expect().
				Status(200).
				Body(Contains(CIDVal)),
		))

		with(testGatewayWithManyProtocols(t,
			"request for {CID}.ipfs.example.com/ipfs/{CID} should return HTTP 404",
			"ensure /ipfs/ namespace is not mounted on subdomain",
			URL("{{scheme}}://{{cid}}.ipfs.{{host}}/ipfs/{{cid}}", u.Scheme, CIDv1, u.Host),
			Expect().
				Status(404),
		))

		with(testGatewayWithManyProtocols(t,
			"request for {CID}.ipfs.example.com/ipfs/file.txt should return data from a file in CID content root",
			"ensure requests to /ipfs/* are not blocked, if content root has such subdirectory",
			URL("{{scheme}}://{{cid}}.ipfs.{{host}}/ipfs/file.txt", u.Scheme, DirCID, u.Host),
			Expect().
				Status(200).
				Body(Contains("I am a txt file")),
		))

		with(testGatewayWithManyProtocols(t,
			"valid file and subdirectory paths in directory listing at {cid}.ipfs.example.com",
			"{CID}.ipfs.example.com/sub/dir (Directory Listing)",
			URL("{{scheme}}://{{cid}}.ipfs.{{host}}/", u.Scheme, DirCID, u.Host),
			Expect().
				Status(200).
				Body(And(
					// TODO: implement html expectations
					Contains(`<a href="/hello">hello</a>`),
					Contains(`<a href="/ipfs">ipfs</a>`),
				)),
		))

		with(testGatewayWithManyProtocols(t,
			"valid parent directory path in directory listing at {cid}.ipfs.example.com/sub/dir",
			"",
			URL("{{scheme}}://{{cid}}.ipfs.{{host}}/ipfs/ipns/", u.Scheme, DirCID, u.Host),
			Expect().
				Status(200).
				Body(And(
					// TODO: implement html expectations
					Contains(`<a href="/ipfs/ipns/..">..</a>`),
					Contains(`<a href="/ipfs/ipns/bar">bar</a>`),
				)),
		))

		with(testGatewayWithManyProtocols(t,
			"request for deep path resource at {cid}.ipfs.localhost/sub/dir/file",
			"",
			URL("{{scheme}}://{{cid}}.ipfs.{{host}}/ipfs/ipns/bar", u.Scheme, DirCID, u.Host),
			Expect().
				Status(200).
				Body(Contains("text-file-content")),
		))

		with(testGatewayWithManyProtocols(t,
			"valid breadcrumb links in the header of directory listing at {cid}.ipfs.example.com/sub/dir",
			`
			Note 1: we test for sneaky subdir names  {cid}.ipfs.example.com/ipfs/ipns/ :^)
			Note 2: example.com/ipfs/.. present in HTML will be redirected to subdomain, so this is expected behavior
			`,
			URL("{{scheme}}://{{cid}}.ipfs.{{host}}/ipfs/ipns/", u.Scheme, DirCID, u.Host),
			Expect().
				Status(200).
				Body(
					And(
						Contains("Index of"),
						Contains(`/ipfs/<a href="//{{host}}/ipfs/{{cid}}">{{cid}}</a>/<a href="//{{host}}/ipfs/{{cid}}/ipfs">ipfs</a>/<a href="//{{host}}/ipfs/{{cid}}/ipfs/ipns">ipns</a>`,
							u.Host, DirCID),
					),
				),
		))

		// TODO: # *.ipns.localhost
		// TODO: # <libp2p-key>.ipns.localhost
		// TODO: # <dnslink-fqdn>.ipns.localhost

		// ## ============================================================================
		// ## Test DNSLink inlining on HTTP gateways
		// ## ============================================================================

		// TODO

		// ## ============================================================================
		// ## Test subdomain-based requests with a custom hostname config
		// ## (origin per content root at http://*.example.com)
		// ## ============================================================================

		// # example.com/ip(f|n)s/*
		// # =============================================================================

		// # path requests to the root hostname should redirect
		// # to a subdomain URL with proper origin isolation

		with(testGatewayWithManyProtocols(t,
			"request for example.com/ipfs/{CIDv1} produces redirect to {CIDv1}.ipfs.example.com",
			"path requests to the root hostname should redirect to a subdomain URL with proper origin isolation",
			URL("{{scheme}}://{{host}}/ipfs/{{cid}}/", u.Scheme, u.Host, CIDv1),
			Expect().
				Headers(
					Header("Location").Equals("{{scheme}}://{{cid}}.ipfs.{{host}}/", u.Scheme, CIDv1, u.Host),
				),
		))

		with(testGatewayWithManyProtocols(t,
			"request for example.com/ipfs/{InvalidCID} produces useful error before redirect",
			"error message should include original CID (and it should be case-sensitive, as we can't assume everyone uses base32)",
			URL("{{scheme}}://{{host}}/ipfs/QmInvalidCID", u.Scheme, u.Host),
			Expect().
				Body(Contains(`invalid path "/ipfs/QmInvalidCID"`)),
		))

		with(testGatewayWithManyProtocols(t,
			"request for example.com/ipfs/{CIDv0} produces redirect to {CIDv1}.ipfs.example.com",
			"",
			URL("{{scheme}}://{{host}}/ipfs/{{cid}}/", u.Scheme, u.Host, CIDv0),
			Expect().
				Status(301).
				Headers(
					Header("Location").Equals("{{scheme}}://{{cid}}.ipfs.{{host}}/", u.Scheme, CIDv0to1, u.Host),
				),
		))

		with(testGatewayWithManyProtocols(t,
			"request for http://example.com/ipfs/{CID} with X-Forwarded-Proto: https produces redirect to HTTPS URL",
			"Support X-Forwarded-Proto",
			Request().
				URL("{{scheme}}://{{host}}/ipfs/{{cid}}/", u.Scheme, u.Host, CIDv1).
				Header("X-Forwarded-Proto", "https"),
			Expect().
				Status(301).
				Headers(
					Header("Location").Equals("https://{{cid}}.ipfs.{{host}}/", CIDv1, u.Host),
				),
		))

		with(testGatewayWithManyProtocols(t,
			"request for example.com/ipfs/?uri=ipfs%3A%2F%2F.. produces redirect to /ipfs/.. content path",
			"Support ipfs:// in https://developer.mozilla.org/en-US/docs/Web/API/Navigator/registerProtocolHandler",
			Request().
				URL("{{scheme}}://{{host}}/ipfs/", u.Scheme, u.Host).
				Query(
					"uri", "ipfs://{{host}}/wiki/Diego_Maradona.html", CIDWikipedia,
				),
			Expect().
				Status(301).
				Headers(
					Header("Location").Equals("/ipfs/{{cid}}/wiki/Diego_Maradona.html", CIDWikipedia),
				),
		))

		// # example.com/ipns/<libp2p-key>
		// TODO

		// # example.com/ipns/<dnslink-fqdn>
		// TODO

		// # DNSLink on Public gateway with a single-level wildcard TLS cert
		// # "Option C" from  https://github.com/ipfs/in-web-browsers/issues/169
		// TODO

		// # Support ipns:// in https://developer.mozilla.org/en-US/docs/Web/API/Navigator/registerProtocolHandler
		// TODO

		// # *.ipns.example.com
		// # ============================================================================

		// # <libp2p-key>.ipns.example.com

		// # API on subdomain gateway example.com
		// # ============================================================================

		// # DNSLink: <dnslink-fqdn>.ipns.example.com
		// # (not really useful outside of localhost, as setting TLS for more than one
		// # level of wildcard is a pain, but we support it if someone really wants it)
		// # ============================================================================
		// TODO

		// # DNSLink on Public gateway with a single-level wildcard TLS cert
		// # "Option C" from  https://github.com/ipfs/in-web-browsers/issues/169

		// ## Test subdomain handling of CIDs that do not fit in a single DNS Label (>63chars)
		// ## https://github.com/ipfs/go-ipfs/issues/7318
		// ## ============================================================================
		// TODO

		with(testGatewayWithManyProtocols(t,
			"request for a too long CID at localhost/ipfs/{CIDv1} returns human readable error",
			"router should not redirect to hostnames that could fail due to DNS limits",
			URL("{{url}}/ipfs/{{cid}}", gatewayURL, CIDv1_TOO_LONG),
			Expect().
				Status(400).
				Body(Contains("CID incompatible with DNS label length limit of 63")),
		))

		with(testGatewayWithManyProtocols(t,
			"request for a too long CID at {CIDv1}.ipfs.localhost returns expected payload",
			"direct request should also fail (provides the same UX as router and avoids confusion)",
			URL("{{scheme}}://{{cid}}.ipfs.{{host}}/", u.Scheme, CIDv1_TOO_LONG, u.Host),
			Expect().
				Status(400).
				Body(Contains("CID incompatible with DNS label length limit of 63")),
		))

		// # public subdomain gateway: *.example.com
		// TODO: IPNS

		// # Disable selected Paths for the subdomain gateway hostname
		// # =============================================================================

		// # disable /ipns for the hostname by not whitelisting it

		// # refuse requests to Paths that were not explicitly whitelisted for the hostname

		// MANY TODOs here

		// ## ============================================================================
		// ## Test support for X-Forwarded-Host
		// ## ============================================================================

		with(testGatewayWithManyProtocols(t,
			"request for http://fake.domain.com/ipfs/{CID} doesn't match the example.com gateway",
			"",
			URL("{{scheme}}://{{domain}}/ipfs/{{cid}}", u.Scheme, "fake.domain.com", CIDv1),
			Expect().
				Status(200),
		))

		with(testGatewayWithManyProtocols(t,
			"request for http://fake.domain.com/ipfs/{CID} with X-Forwarded-Host: example.com match the example.com gateway",
			"",
			Request().
				URL("{{scheme}}://{{domain}}/ipfs/{{cid}}", u.Scheme, "fake.domain.com", CIDv1).
				Header("X-Forwarded-Host", u.Host),
			Expect().
				Status(301).
				Headers(
					Header("Location").Equals("{{scheme}}://{{cid}}.ipfs.{{host}}/", u.Scheme, CIDv1, u.Host),
				),
		))

		with(testGatewayWithManyProtocols(t,
			"request for http://fake.domain.com/ipfs/{CID} with X-Forwarded-Host: example.com and X-Forwarded-Proto: https match the example.com gateway, redirect with https",
			"",
			Request().
				URL("{{scheme}}://{{domain}}/ipfs/{{cid}}", u.Scheme, "fake.domain.com", CIDv1).
				Header("X-Forwarded-Host", u.Host).
				Header("X-Forwarded-Proto", "https"),
			Expect().
				Status(301).
				Headers(
					Header("Location").Equals("https://{{cid}}.ipfs.{{host}}/", CIDv1, u.Host),
				),
		))
	}

	if specs.SubdomainGateway.IsEnabled() {
		Run(t, tests)
	} else {
		t.Skip("subdomain gateway disabled")
	}
}

func testGatewayWithManyProtocols(t *testing.T, label string, hint string, reqURL interface{}, expected ExpectBuilder) SugarTests {
	t.Helper()

	baseURL := ""
	baseReq := Request()

	switch req := reqURL.(type) {
	case string:
		baseURL = reqURL.(string)
	case RequestBuilder:
		baseReq = req
		baseURL = req.GetURL()
	default:
		t.Fatalf("invalid type for reqURL: %T", reqURL)
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		t.Fatal(err)
	}
	// Because you might be testing an IPFS node in CI, or on your local machine, the test are designed
	// to test the subdomain behavior (querying http://{CID}.my-subdomain-gateway.io/) even if the node is
	// actually living on http://127.0.0.1:8080 or somewhere else.
	//
	// The test knows two addresses:
	// 		- GatewayURL: the URL we connect to, it might be "dweb.link", "127.0.0.1:8080", etc.
	// 		- SubdomainGatewayURL: the URL we test for subdomain requests, it might be "dweb.link", "localhost", "example.com", etc.

	// host is the hostname of the gateway we are testing, it might be `localhost` or `example.com`
	host := u.Host

	// raw url is the url but we replace the host with our local url, it might be `http://127.0.0.1/ipfs/something`
	u.Host = GatewayHost
	rawURL := u.String()

	return SugarTests{
		{
			Name: fmt.Sprintf("%s (direct HTTP)", label),
			Hint: fmt.Sprintf("%s\n%s", hint, "direct HTTP request (hostname in URL, raw IP in Host header)"),
			Request: baseReq.
				URL(rawURL).
				DoNotFollowRedirects().
				Headers(
					Header("Host", host),
				),
			Response: expected,
		},
		{
			Name: fmt.Sprintf("%s (HTTP proxy)", label),
			Hint: fmt.Sprintf("%s\n%s", hint, "HTTP proxy (hostname is passed via URL)"),
			Request: baseReq.
				URL(baseURL).
				Proxy(GatewayURL).
				DoNotFollowRedirects(),
			Response: expected,
		},
		{
			Name: fmt.Sprintf("%s (HTTP proxy tunneling via CONNECT)", label),
			Hint: fmt.Sprintf("%s\n%s", hint, `HTTP proxy
				In HTTP/1.x, the pseudo-method CONNECT,
				can be used to convert an HTTP connection into a tunnel to a remote host
				https://tools.ietf.org/html/rfc7231#section-4.3.6
			`),
			Request: baseReq.
				URL(baseURL).
				Proxy(GatewayURL).
				WithProxyTunnel().
				DoNotFollowRedirects().
				Headers(
					Header("Host", host),
				),
			Response: expected,
		},
	}
}
