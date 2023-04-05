package tests

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/dnslink"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestRedirectsFileSupport(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0109-redirects.car")
	redirectDir := fixture.MustGetNode("examples")
	redirectDirCID := redirectDir.Base32Cid()

	// CUSTOM_4XX_CID=$(ipfs resolve -r /ipfs/$CAR_ROOT_CID/examples/404.html | cut -d "/" -f3)
	custom404 := fixture.MustGetNode("examples", "404.html")
	// CUSTOM_4XX_CID=$(ipfs resolve -r /ipfs/$CAR_ROOT_CID/examples/410.html | cut -d "/" -f3)
	custom410 := fixture.MustGetNode("examples", "410.html")
	// CUSTOM_4XX_CID=$(ipfs resolve -r /ipfs/$CAR_ROOT_CID/examples/451.html | cut -d "/" -f3)
	custom451 := fixture.MustGetNode("examples", "451.html")

	tests := SugarTests{}

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

		redirectDirBaseURL := fmt.Sprintf("%s://%s.ipfs.%s", u.Scheme, redirectDirCID, u.Host)

		tests = append(tests, SugarTests{
			// test_expect_success "request for $REDIRECTS_DIR_HOSTNAME/redirect-one redirects with default of 301, per _redirects file" '
			//
			//	curl -sD - --resolve $REDIRECTS_DIR_HOSTNAME:127.0.0.1 "http://$REDIRECTS_DIR_HOSTNAME/redirect-one" > response &&
			//	test_should_contain "301 Moved Permanently" response &&
			//	test_should_contain "Location: /one.html" response
			//
			// '
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/redirect-one redirects with default of 301, per _redirects file",
				Request: Request().
					DoNotFollowRedirects().
					Header("Host", u.Host).
					URL("%s/redirect-one", redirectDirBaseURL),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("/one.html"),
					),
			},
			// test_expect_success "request for $REDIRECTS_DIR_HOSTNAME/301-redirect-one redirects with 301, per _redirects file" '
			//
			//	curl -sD - --resolve $REDIRECTS_DIR_HOSTNAME:127.0.0.1 "http://$REDIRECTS_DIR_HOSTNAME/301-redirect-one" > response &&
			//	test_should_contain "301 Moved Permanently" response &&
			//	test_should_contain "Location: /one.html" response
			//
			// '
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/301-redirect-one redirects with 301, per _redirects file",
				Request: Request().
					DoNotFollowRedirects().
					URL("%s/301-redirect-one", redirectDirBaseURL),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("/one.html"),
					),
			},
			// test_expect_success "request for $REDIRECTS_DIR_HOSTNAME/302-redirect-two redirects with 302, per _redirects file" '
			//
			//	curl -sD - --resolve $REDIRECTS_DIR_HOSTNAME:127.0.0.1 "http://$REDIRECTS_DIR_HOSTNAME/302-redirect-two" > response &&
			//	test_should_contain "302 Found" response &&
			//	test_should_contain "Location: /two.html" response
			//
			// '
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/302-redirect-two redirects with 302, per _redirects file",
				Request: Request().
					DoNotFollowRedirects().
					URL("%s/302-redirect-two", redirectDirBaseURL),
				Response: Expect().
					Status(302).
					Headers(
						Header("Location").Equals("/two.html"),
					),
			},
			// test_expect_success "request for $REDIRECTS_DIR_HOSTNAME/200-index returns 200, per _redirects file" '
			//
			//	curl -sD - --resolve $REDIRECTS_DIR_HOSTNAME:127.0.0.1 "http://$REDIRECTS_DIR_HOSTNAME/200-index" > response &&
			//	test_should_contain "my index" response &&
			//	test_should_contain "200 OK" response
			//
			// '
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/200-index returns 200, per _redirects file",
				Request: Request().
					DoNotFollowRedirects().
					URL("%s/200-index", redirectDirBaseURL),
				Response: Expect().
					Status(200).
					Body(Contains("my index")),
			},
			// test_expect_success "request for $REDIRECTS_DIR_HOSTNAME/posts/:year/:month/:day/:title redirects with 301 and placeholders, per _redirects file" '
			//
			//	curl -sD - --resolve $REDIRECTS_DIR_HOSTNAME:127.0.0.1 "http://$REDIRECTS_DIR_HOSTNAME/posts/2022/01/01/hello-world" > response &&
			//	test_should_contain "301 Moved Permanently" response &&
			//	test_should_contain "Location: /articles/2022/01/01/hello-world" response
			//
			// '
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/posts/:year/:month/:day/:title redirects with 301 and placeholders, per _redirects file",
				Request: Request().
					DoNotFollowRedirects().
					URL("%s/posts/2022/01/01/hello-world", redirectDirBaseURL),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("/articles/2022/01/01/hello-world"),
					),
			},
			// test_expect_success "request for $REDIRECTS_DIR_HOSTNAME/splat/one.html redirects with 301 and splat placeholder, per _redirects file" '
			//
			//	curl -sD - --resolve $REDIRECTS_DIR_HOSTNAME:127.0.0.1 "http://$REDIRECTS_DIR_HOSTNAME/splat/one.html" > response &&
			//	test_should_contain "301 Moved Permanently" response &&
			//	test_should_contain "Location: /redirected-splat/one.html" response
			//
			// '
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/splat/one.html redirects with 301 and splat placeholder, per _redirects file",
				Request: Request().
					DoNotFollowRedirects().
					URL("%s/splat/one.html", redirectDirBaseURL),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("/redirected-splat/one.html"),
					),
			},
			// # ensure custom 4xx works and has the same cache headers as regular /ipfs/ path
			// test_expect_success "request for $REDIRECTS_DIR_HOSTNAME/not-found/has-no-redirects-entry returns custom 404, per _redirects file" '
			//
			//	curl -sD - --resolve $REDIRECTS_DIR_HOSTNAME:127.0.0.1 "http://$REDIRECTS_DIR_HOSTNAME/not-found/has-no-redirects-entry" > response &&
			//	test_should_contain "404 Not Found" response &&
			//	test_should_contain "Cache-Control: public, max-age=29030400, immutable" response &&
			//	test_should_contain "Etag: \"$CUSTOM_4XX_CID\"" response &&
			//	test_should_contain "my 404" response
			//
			// '
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/not-found/has-no-redirects-entry returns custom 404, per _redirects file",
				Request: Request().
					URL("%s/not-found/has-no-redirects-entry", redirectDirBaseURL),
				Response: Expect().
					Status(404).
					Headers(
						Header("Cache-Control").Equals("public, max-age=29030400, immutable"),
						Header("Etag").Equals("\"%s\"", custom404.Cid().String()),
					).
					Body(Contains(custom404.ReadFile())),
			},
			// CUSTOM_4XX_CID=$(ipfs resolve -r /ipfs/$CAR_ROOT_CID/examples/410.html | cut -d "/" -f3)
			// test_expect_success "request for $REDIRECTS_DIR_HOSTNAME/gone/has-no-redirects-entry returns custom 410, per _redirects file" '
			//   curl -sD - --resolve $REDIRECTS_DIR_HOSTNAME:127.0.0.1 "http://$REDIRECTS_DIR_HOSTNAME/gone/has-no-redirects-entry" > response &&
			//   test_should_contain "410 Gone" response &&
			//   test_should_contain "Cache-Control: public, max-age=29030400, immutable" response &&
			//   test_should_contain "Etag: \"$CUSTOM_4XX_CID\"" response &&
			//   test_should_contain "my 410" response
			// '
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/gone/has-no-redirects-entry returns custom 410, per _redirects file",
				Request: Request().
					URL("%s/gone/has-no-redirects-entry", redirectDirBaseURL),
				Response: Expect().
					Status(410).
					Headers(
						Header("Cache-Control").Equals("public, max-age=29030400, immutable"),
						Header("Etag").Equals("\"%s\"", custom410.Cid().String()),
					).
					Body(Contains(custom410.ReadFile())),
			},
			// CUSTOM_4XX_CID=$(ipfs resolve -r /ipfs/$CAR_ROOT_CID/examples/451.html | cut -d "/" -f3)
			// test_expect_success "request for $REDIRECTS_DIR_HOSTNAME/unavail/has-no-redirects-entry returns custom 451, per _redirects file" '
			//
			//	curl -sD - --resolve $REDIRECTS_DIR_HOSTNAME:127.0.0.1 "http://$REDIRECTS_DIR_HOSTNAME/unavail/has-no-redirects-entry" > response &&
			//	test_should_contain "451 Unavailable For Legal Reasons" response &&
			//	test_should_contain "Cache-Control: public, max-age=29030400, immutable" response &&
			//	test_should_contain "Etag: \"$CUSTOM_4XX_CID\"" response &&
			//	test_should_contain "my 451" response
			//
			// '
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/unavail/has-no-redirects-entry returns custom 451, per _redirects file",
				Request: Request().
					URL("%s/unavail/has-no-redirects-entry", redirectDirBaseURL),
				Response: Expect().
					Status(451).
					Headers(
						Header("Cache-Control").Equals("public, max-age=29030400, immutable"),
						Header("Etag").Equals("\"%s\"", custom451.Cid().String()),
					).
					Body(Contains(custom451.ReadFile())),
			},
			// test_expect_success "request for $REDIRECTS_DIR_HOSTNAME/catch-all returns 200, per _redirects file" '
			//
			//	curl -sD - --resolve $REDIRECTS_DIR_HOSTNAME:127.0.0.1 "http://$REDIRECTS_DIR_HOSTNAME/catch-all" > response &&
			//	test_should_contain "200 OK" response &&
			//	test_should_contain "my index" response
			//
			// '
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/catch-all returns 200, per _redirects file",
				Request: Request().
					URL("%s/catch-all", redirectDirBaseURL),
				Response: Expect().
					Status(200).
					Body(Contains("my index")),
			},
			// # This test ensures _redirects is supported only on Web Gateways that use Host header (DNSLink, Subdomain)
			// test_expect_success "request for http://127.0.0.1:$GWAY_PORT/ipfs/$REDIRECTS_DIR_CID/301-redirect-one returns generic 404 (no custom 404 from _redirects since no origin isolation)" '
			//
			//	curl -sD - "http://127.0.0.1:$GWAY_PORT/ipfs/$REDIRECTS_DIR_CID/301-redirect-one" > response &&
			//	test_should_contain "404 Not Found" response &&
			//	test_should_not_contain "my 404" response
			//
			// '
			{
				// TODO: how to test this correctly?
				Name: "This test ensures _redirects is supported only on Web Gateways that use Host header (DNSLink, Subdomain)",
				Hint: `
			We expect the request to fail with a 404 (do not use the _redirect), and that 404 should not contain the custom 404 body.
			`,
				Request: Request().
					URL("http://127.0.0.1:8080/ipfs/%s/301-redirect-one", redirectDirCID),
				Response: Expect().
					Status(404).
					Body(Not(Contains(custom404.ReadFile()))),
			},
		}...)

		// # Invalid file, containing forced redirect
		// INVALID_REDIRECTS_DIR_CID=$(ipfs resolve -r /ipfs/$CAR_ROOT_CID/forced | cut -d "/" -f3)
		invalidRedirectsDirCID := fixture.MustGetNode("forced").Base32Cid()
		// INVALID_REDIRECTS_DIR_HOSTNAME="${INVALID_REDIRECTS_DIR_CID}.ipfs.localhost:$GWAY_PORT"
		invalidDirBaseURL := fmt.Sprintf("%s://%s.ipfs.%s", u.Scheme, invalidRedirectsDirCID, u.Host)
		// TOO_LARGE_REDIRECTS_DIR_CID=$(ipfs resolve -r /ipfs/$CAR_ROOT_CID/too-large | cut -d "/" -f3)
		tooLargeRedirectsDirCID := fixture.MustGetNode("too-large").Base32Cid()
		// TOO_LARGE_REDIRECTS_DIR_HOSTNAME="${TOO_LARGE_REDIRECTS_DIR_CID}.ipfs.localhost:$GWAY_PORT"
		tooLargeDirBaseURL := fmt.Sprintf("%s://%s.ipfs.%s", u.Scheme, tooLargeRedirectsDirCID, u.Host)

		tests = append(tests, SugarTests{
			// # if accessing a path that doesn't exist, read _redirects and fail parsing, and return error
			// test_expect_success "invalid file: request for $INVALID_REDIRECTS_DIR_HOSTNAME/not-found returns error about invalid redirects file" '
			//   curl -sD - --resolve $INVALID_REDIRECTS_DIR_HOSTNAME:127.0.0.1 "http://$INVALID_REDIRECTS_DIR_HOSTNAME/not-found" > response &&
			//   test_should_contain "500" response &&
			//   test_should_contain "could not parse _redirects:" response &&
			//   test_should_contain "forced redirects (or \"shadowing\") are not supported" response
			// '
			{
				Name: "invalid file: request for $INVALID_REDIRECTS_DIR_HOSTNAME/not-found returns error about invalid redirects file",
				Hint: `if accessing a path that doesn't exist, read _redirects and fail parsing, and return error`,
				Request: Request().
					URL("%s/not-found", invalidDirBaseURL),
				Response: Expect().
					Status(500).
					Body(
						And(
							Contains("could not parse _redirects:"),
							Contains("forced redirects (or \"shadowing\") are not supported"),
						),
					),
			},
			// # if accessing a path that doesn't exist and _redirects file is too large, return error
			// test_expect_success "invalid file: request for $TOO_LARGE_REDIRECTS_DIR_HOSTNAME/not-found returns error about too large redirects file" '
			//   curl -sD - --resolve $TOO_LARGE_REDIRECTS_DIR_HOSTNAME:127.0.0.1 "http://$TOO_LARGE_REDIRECTS_DIR_HOSTNAME/not-found" > response &&
			//   test_should_contain "500" response &&
			//   test_should_contain "could not parse _redirects:" response &&
			//   test_should_contain "redirects file size cannot exceed" response
			// '
			{
				Name: "invalid file: request for $TOO_LARGE_REDIRECTS_DIR_HOSTNAME/not-found returns error about too large redirects file",
				Hint: `if accessing a path that doesn't exist and _redirects file is too large, return error`,
				Request: Request().
					URL("%s/not-found", tooLargeDirBaseURL),
				Response: Expect().
					Status(500).
					Body(
						And(
							Contains("could not parse _redirects:"),
							Contains("redirects file size cannot exceed"),
						),
					),
			},
		}...)
	}

	if specs.SubdomainGateway.IsEnabled() {
		Run(t, unwrapTests(t, tests))
	} else {
		t.Skip("subdomain gateway disabled")
	}
}

func TestRedirectsFileSupportWithDNSLink(t *testing.T) {
	dnsLinks := dnslink.MustOpenDNSLink("t0109-dnslink.yml")
	dnsLink := dnsLinks.Get("custom-dnslink")

	gatewayURL := SubdomainGatewayURL
	u, err := url.Parse(gatewayURL)
	if err != nil {
		t.Fatal(err)
	}

	dnsLinkBaseUrl := fmt.Sprintf("%s://%s.%s", u.Scheme, dnsLink, u.Host)

	tests := SugarTests{
		// # make sure test setup is valid (fail if CoreAPI is unable to resolve)
		// test_expect_success "spoofed DNSLink record resolves in cli" "
		// ipfs resolve /ipns/$DNSLINK_FQDN > result &&
		// test_should_contain \"$REDIRECTS_DIR_CID\" result &&
		// ipfs cat /ipns/$DNSLINK_FQDN/_redirects > result &&
		// test_should_contain \"index.html\" result
		// "
		// SKIPPED

		// test_expect_success "request for $DNSLINK_FQDN/redirect-one redirects with default of 301, per _redirects file" '
		// curl -sD - --resolve $DNSLINK_FQDN:$GWAY_PORT:127.0.0.1 "http://$DNSLINK_FQDN:$GWAY_PORT/redirect-one" > response &&
		// test_should_contain "301 Moved Permanently" response &&
		// test_should_contain "Location: /one.html" response
		// '
		{
			Name: "request for $DNSLINK_FQDN/redirect-one redirects with default of 301, per _redirects file",
			Request: Request().
				URL("%s/redirect-one", dnsLinkBaseUrl),
			Response: Expect().
				Status(301).
				Headers(
					Header("Location", "/one.html"),
				),
		},
		// # ensure custom 404 works and has the same cache headers as regular /ipns/ paths
		// test_expect_success "request for $DNSLINK_FQDN/en/has-no-redirects-entry returns custom 404, per _redirects file" '
		// curl -sD - --resolve $DNSLINK_FQDN:$GWAY_PORT:127.0.0.1 "http://$DNSLINK_FQDN:$GWAY_PORT/not-found/has-no-redirects-entry" > response &&
		// test_should_contain "404 Not Found" response &&
		// test_should_contain "Etag: \"Qmd9GD7Bauh6N2ZLfNnYS3b7QVAijbud83b8GE8LPMNBBP\"" response &&
		// test_should_not_contain "Cache-Control: public, max-age=29030400, immutable" response &&
		// test_should_not_contain "immutable" response &&
		// test_should_contain "Date: " response &&
		// test_should_contain "my 404" response
		// '
		{
			Name: "request for $DNSLINK_FQDN/en/has-no-redirects-entry returns custom 404, per _redirects file",
			Hint: `ensure custom 404 works and has the same cache headers as regular /ipns/ paths`,
			Request: Request().
				URL("%s/not-found/has-no-redirects-entry", dnsLinkBaseUrl),
			Response: Expect().
				Status(404).
				Headers(
					Header("Etag", "\"Qmd9GD7Bauh6N2ZLfNnYS3b7QVAijbud83b8GE8LPMNBBP\""),
					Header("Cache-Control").Not().Contains("public, max-age=29030400, immutable"),
					Header("Cache-Control").Not().Contains("immutable"),
					Header("Date").Exists(),
				).
				Body(
					// TODO: I like the readable part here, maybe rewrite to load the file.
					Contains("my 404"),
				),
		},
		// test_expect_success "request for $NO_DNSLINK_FQDN/redirect-one does not redirect, since DNSLink is disabled" '
		// curl -sD - --resolve $NO_DNSLINK_FQDN:$GWAY_PORT:127.0.0.1 "http://$NO_DNSLINK_FQDN:$GWAY_PORT/redirect-one" > response &&
		// test_should_not_contain "one.html" response &&
		// test_should_not_contain "301 Moved Permanently" response &&
		// test_should_not_contain "Location:" response
		// '
		// TODO(lidel): this test seems to validate some kubo behavior not really gateway.
		// {
		// 	Name: "request for $NO_DNSLINK_FQDN/redirect-one does not redirect, since DNSLink is disabled",
		// 	Request: Request().
		// 		URL("%s/redirect-one", noDnsLinkBaseUrl),
		// 	Response: Expect().
		// 		// TODO: add "status not equal to 301" check.
		// 		// TODO: what `test_should_not_contain "one.html" response` actually means? No location correct?
		// 		Headers(
		// 			Header("Location").Not().Exists(),
		// 		),
		// },
	}

	if specs.DNSLinkResolver.IsEnabled() {
		Run(t, unwrapTests(t, tests))
	} else {
		t.Skip("subdomain gateway disabled")
	}
}

func unwrapTests(t *testing.T, tests SugarTests) SugarTests {
	t.Helper()

	var out SugarTests
	for _, test := range tests {
		out = append(out, unwrapTestForGateway(t, test)...)
	}
	return out
}

func unwrapTestForGateway(t *testing.T, test SugarTest) SugarTests {
	t.Helper()

	baseURL := test.Request.GetURL()
	req := test.Request
	expected := test.Response

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
			Name: fmt.Sprintf("%s (direct HTTP)", test.Name),
			Hint: fmt.Sprintf("%s\n%s", test.Hint, "direct HTTP request (hostname in URL, raw IP in Host header)"),
			Request: req.
				URL(rawURL).
				DoNotFollowRedirects().
				Headers(
					Header("Host", host),
				),
			Response: expected,
		},
		{
			Name: fmt.Sprintf("%s (HTTP proxy)", test.Name),
			Hint: fmt.Sprintf("%s\n%s", test.Hint, "HTTP proxy (hostname is passed via URL)"),
			Request: req.
				URL(baseURL).
				Proxy(GatewayURL).
				DoNotFollowRedirects(),
			Response: expected,
		},
		{
			Name: fmt.Sprintf("%s (HTTP proxy tunneling via CONNECT)", test.Name),
			Hint: fmt.Sprintf("%s\n%s", test.Hint, `HTTP proxy
				In HTTP/1.x, the pseudo-method CONNECT,
				can be used to convert an HTTP connection into a tunnel to a remote host
				https://tools.ietf.org/html/rfc7231#section-4.3.6
			`),
			Request: req.
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
