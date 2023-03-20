package tests

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	. "github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewaySubdomains(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0114-gateway_subdomains.car")

	CIDVal := "hello" // TODO
	DirCID := fixture.MustGetCid("testdirlisting")
	CIDv1 := fixture.MustGetCid("hello-CIDv1")
	CIDv0 := fixture.MustGetCid("hello-CIDv0")
	CIDv0to1 := fixture.MustGetCid("hello-CIDv0to1")
	CIDv1_TOO_LONG := fixture.MustGetCid("hello-CIDv1_TOO_LONG")

	fmt.Println("DIR_CID:", DirCID)
	fmt.Println("CIDv1:", CIDv1, "CIDv0:", CIDv0)
	fmt.Println("CIDv0to1:", CIDv0to1, "CIDv1_TOO_LONG:", CIDv1_TOO_LONG)

	tests := []CTest{}

	with := func(moreTests []CTest) {
		tests = append(tests, moreTests...)
	}

	gatewayURLs := []string{
		SubdomainGatewayUrl,
		SubdomainLocalhostGatewayUrl,
	}

	with(testLocalhostGatewayResponseShouldContain(t,
		"request for 127.0.0.1/ipfs/{CID} stays on path",
		// TODO: this feature should get disabled if we're not taling with an IP gateway.
		`IP remains old school path-based gateway`,
		fmt.Sprintf("%s/ipfs/%s/", GatewayUrl, CIDv1),
		Expect().
			Status(200).
			Body("hello\n").
			Response(),
	))

	for _, gatewayURL := range gatewayURLs {
		u, err := url.Parse(gatewayURL)
		if err != nil {
			t.Fatal(err)
		}

		with(testLocalhostGatewayResponseShouldContain(t,
			"request for example.com/ipfs/{CIDv1} redirects to subdomain",
			`
			subdomains should not return payload directly,
			but redirect to URL with proper origin isolation
			`,
			fmt.Sprintf("%s/ipfs/%s/", gatewayURL, CIDv1),
			Expect().
				Status(301).
				Headers(
					Header("Location").
						Hint("request for example.com/ipfs/{CIDv1} returns Location HTTP header for subdomain redirect in browsers").
						Contains("%s://%s.ipfs.%s/", u.Scheme, CIDv1, u.Host),
				).
				// We return body with HTTP 301 so existing cli scripts that use path-based
				// gateway do not break (curl doesn't auto-redirect without passing -L; wget
				// does not span across hostnames by default)
				// Context: https://github.com/ipfs/go-ipfs/issues/6975
				Body("hello\n").
				Response(),
		))

		with(testLocalhostGatewayResponseShouldContain(t,
			"request for example.com/ipfs/{DirCID} redirects to subdomain",
			`
			subdomains should not return payload directly,
			but redirect to URL with proper origin isolation
			`,
			fmt.Sprintf("%s/ipfs/%s/", gatewayURL, DirCID),
			Expect().
				Status(301).
				Headers(
					Header("Location").
						Hint("request for example.com/ipfs/{DirCID} returns Location HTTP header for subdomain redirect in browsers").
						Contains("%s://%s.ipfs.%s/", u.Scheme, DirCID, u.Host),
				).Response(),
		))

		with(testLocalhostGatewayResponseShouldContain(t,
			"request for example.com/ipfs/{CIDv0} redirects to CIDv1 representation in subdomain",
			"",
			fmt.Sprintf("%s/ipfs/%s/", gatewayURL, CIDv0),
			Expect().
				Status(301).
				Headers(
					Header("Location").
						Hint("request for example.com/ipfs/{CIDv0to1} returns Location HTTP header for subdomain redirect in browsers").
						Contains("%s://%s.ipfs.%s/", u.Scheme, CIDv0to1, u.Host),
				).Response(),
		))

		// TODO: ipns
		// TODO: dns link test

		// API on localhost subdomain gateway

		with(testLocalhostGatewayResponseShouldContain(t,
			"/api/v0 present on the root hostname",
			"request for localhost/api",
			fmt.Sprintf("%s/api/v0/refs?arg=%s&r=true", gatewayURL, DirCID),
			Expect().
				Status(200).
				Body(Contains(
					"Ref",
				)).Response(),
		))

		with(testLocalhostGatewayResponseShouldContain(t,
			"/api/v0 not mounted on content root subdomains",
			"request for {cid}.ipfs.examle.com/api returns data if present on the content root",
			fmt.Sprintf("%s://%s.ipfs.%s/api/file.txt", u.Scheme, DirCID, u.Host),
			Expect().
				Status(200).
				Body(Contains(
					"I am a txt file",
				)).Response(),
		))

		with(testLocalhostGatewayResponseShouldContain(t,
			"request for {cid}.ipfs.localhost/api/v0/refs returns 404",
			"",
			fmt.Sprintf("%s://%s.ipfs.%s/api/v0/refs?arg=%s&r=true", u.Scheme, DirCID, u.Host, DirCID),
			Expect().
				Status(404).
				Response(),
		))

		// ============================================================================
		// Test subdomain-based requests to a local gateway with default config
		// (origin per content root at http://*.localhost)
		// ============================================================================

		with(testLocalhostGatewayResponseShouldContain(t,
			"request for {CID}.ipfs.example.com should return expected payload",
			"",
			fmt.Sprintf("%s://%s.ipfs.%s", u.Scheme, CIDv1, u.Host),
			Expect().
				Status(200).
				Body(Contains(CIDVal)).
				Response(),
		))

		with(testLocalhostGatewayResponseShouldContain(t,
			"request for {CID}.ipfs.example.com/ipfs/{CID} should return HTTP 404",
			"ensure /ipfs/ namespace is not mounted on subdomain",
			fmt.Sprintf("%s://%s.ipfs.%s/ipfs/%s", u.Scheme, CIDv1, u.Host, CIDv1),
			Expect().
				Status(404).
				Response(),
		))

		with(testLocalhostGatewayResponseShouldContain(t,
			"request for {CID}.ipfs.example.com/ipfs/file.txt should return data from a file in CID content root",
			"ensure requests to /ipfs/* are not blocked, if content root has such subdirectory",
			fmt.Sprintf("%s://%s.ipfs.%s/ipfs/file.txt", u.Scheme, DirCID, u.Host),
			Expect().
				Status(200).
				Body(Contains("I am a txt file")).
				Response(),
		))

		with(testLocalhostGatewayResponseShouldContain(t,
			"valid file and subdirectory paths in directory listing at {cid}.ipfs.localhost",
			"{CID}.ipfs.example.com/sub/dir (Directory Listing)",
			fmt.Sprintf("%s://%s.ipfs.%s/", u.Scheme, DirCID, u.Host),
			Expect().
				Status(200).
				Body(And(
					// TODO: implement html expectations
					Contains("<a href=\"/hello\">hello</a>"),
					Contains("<a href=\"/ipfs\">ipfs</a>"),
				)).
				Response(),
		))

		with(testLocalhostGatewayResponseShouldContain(t,
			"valid parent directory path in directory listing at {cid}.ipfs.localhost/sub/dir",
			"",
			fmt.Sprintf("%s://%s.ipfs.%s/ipfs/ipns/", u.Scheme, DirCID, u.Host),
			Expect().
				Status(200).
				Body(And(
					// TODO: implement html expectations
					Contains("<a href=\"/ipfs/ipns/..\">..</a>"),
					Contains("<a href=\"/ipfs/ipns/bar\">bar</a>"),
				)).
				Response(),
		))

		with(testLocalhostGatewayResponseShouldContain(t,
			"request for deep path resource at {cid}.ipfs.localhost/sub/dir/file",
			"",
			fmt.Sprintf("%s://%s.ipfs.%s/ipfs/ipns/bar", u.Scheme, DirCID, u.Host),
			Expect().
				Status(200).
				Body(Contains("text-file-content")).
				Response(),
		))

		// TODO: # *.ipns.localhost
		// TODO: # <libp2p-key>.ipns.localhost
		// TODO: # <dnslink-fqdn>.ipns.localhost

		// # api.localhost/api

		with(testLocalhostGatewayResponseShouldContain(t,
			"request for api.localhost returns API response",
			"Note: we use DIR_CID so refs -r returns some CIDs for child nodes",
			fmt.Sprintf("%s://api.%s/api/v0/refs?arg=%s&r=true", u.Scheme, u.Host, DirCID),
			Expect().
				Status(200).
				Body(Contains("Ref")).
				Response(),
		))

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

		// TODO: lidel - for some reason in 114, we only test the simple case, with Host, no proxy or tunnel.
		// Do we need to port `test_hostname_gateway_response_should_contain` or can we
		// reuse the same "larger" test.
		with(testLocalhostGatewayResponseShouldContain(t,
			"request for example.com/ipfs/{CIDv1} produces redirect to {CIDv1}.ipfs.example.com",
			"path requests to the root hostname should redirect to a subdomain URL with proper origin isolation",
			fmt.Sprintf("%s://%s/ipfs/%s/", u.Scheme, u.Host, CIDv1),
			Expect().
			Headers(
					Header("Location").Equals("%s://%s.ipfs.%s/", u.Scheme, CIDv1, u.Host),
				).
				Response(),
		))

		with(testLocalhostGatewayResponseShouldContain(t,
			"request for example.com/ipfs/{InvalidCID} produces useful error before redirect",
			"error message should include original CID (and it should be case-sensitive, as we can't assume everyone uses base32)",
			fmt.Sprintf("%s://%s/ipfs/QmInvalidCID", u.Scheme, u.Host),
			Expect().
				Body(Contains("invalid path \"/ipfs/QmInvalidCID\"")).
				Response(),
		))

		with(testLocalhostGatewayResponseShouldContain(t,
			"request for example.com/ipfs/{CIDv0} produces redirect to {CIDv1}.ipfs.example.com",
			"",
			fmt.Sprintf("%s://%s/ipfs/%s/", u.Scheme, u.Host, CIDv0),
			Expect().
				Status(301).
				Headers(
					Header("Location").Equals("%s://%s.ipfs.%s/", u.Scheme, CIDv0to1, u.Host),
				).
				Response(),
		))

		with(testLocalhostGatewayResponseShouldContain(t,
			"request for http://example.com/ipfs/{CID} with X-Forwarded-Proto: https produces redirect to HTTPS URL",
			"Support X-Forwarded-Proto",
			Request().
				URL("%s://%s/ipfs/%s/", u.Scheme, u.Host, CIDv1).
				Header("X-Forwarded-Proto", "https"),
			Expect().
				Status(301).
				Headers(
					Header("Location").Equals("https://%s.ipfs.%s/", CIDv1, u.Host),
				).
				Response(),
		))

		// # Support ipfs:// in https://developer.mozilla.org/en-US/docs/Web/API/Navigator/registerProtocolHandler
		// TODO
	}

	if SubdomainGateway.IsEnabled() {
		Run(t, tests)
	}
}

func testLocalhostGatewayResponseShouldContain(t *testing.T, label string, hint string, reqUrl interface{}, expected CResponse) []CTest {
	t.Helper()

	baseUrl := ""
	baseReq := Request()

	switch req := reqUrl.(type) {
	case string:
		baseUrl = reqUrl.(string)
	case RequestBuilder:
		baseReq = req
		// TODO: we use the Request builder to easily move URL and headers around and make them reusable.
		// This probably deserves a better API that throws if you passed a URL with just a path (we don't resolve relative URLs in the builder).
		baseUrl = req.URL_
	default:
		t.Fatalf("invalid type for reqUrl: %T", reqUrl)
	}

	u, err := url.Parse(baseUrl)
	if err != nil {
		t.Fatal(err)
	}

	// Careful: host and hostname are reversed in the sharness t0114 test.
	// host := u.Host // not used, because we don't need the proxy.
	host := u.Host

	// proxy is the base url we use in the test suite

	// raw url is the url but we replace the host with our local url
	u.Host = GatewayHost
	rawUrl := u.String()

	return []CTest{
		{
			Name: fmt.Sprintf("%s (direct HTTP)", label),
			Hint: fmt.Sprintf("%s\n%s", hint, "direct HTTP request (hostname in URL, raw IP in Host header)"),
			Request: baseReq.
				URL(rawUrl).
				DoNotFollowRedirects().
				Headers(
					Header("Host", host),
				).
				Request(),
			Response: expected,
		},
		{
			Name: fmt.Sprintf("%s (HTTP proxy)", label),
			Hint: fmt.Sprintf("%s\n%s", hint, "HTTP proxy (hostname is passed via URL)"),
			Request: baseReq.
				URL(baseUrl).
				Proxy(GatewayUrl).
				DoNotFollowRedirects().
				Request(),
			Response: expected,
		},
		// TODO: port the proxy1.0 as well (use http 1) but also http2?
		{
			Name: fmt.Sprintf("%s (HTTP proxy tunneling via CONNECT)", label),
			Hint: fmt.Sprintf("%s\n%s", hint, `HTTP proxy
				In HTTP/1.x, the pseudo-method CONNECT,
				can be used to convert an HTTP connection into a tunnel to a remote host
				https://tools.ietf.org/html/rfc7231#section-4.3.6
			`),
			Request: baseReq.
				URL(baseUrl).
				Proxy(GatewayUrl).
				WithProxyTunnel().
				DoNotFollowRedirects().
				Headers(
					Header("Host", host),
				).
				Request(),
			Response: expected,
		},
	}
}
