package tests

import (
	"fmt"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestRedirectsFileSupport(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0109-redirects.car")

	redirectDir := fixture.MustGetNode("examples")
	redirectDirCID := redirectDir.Base32Cid()

	redirectDirHostname := fmt.Sprintf("%s.ipfs.localhost:8080", redirectDirCID)

	// CUSTOM_4XX_CID=$(ipfs resolve -r /ipfs/$CAR_ROOT_CID/examples/404.html | cut -d "/" -f3)
	custom404 := fixture.MustGetNode("examples", "404.html")
	// CUSTOM_4XX_CID=$(ipfs resolve -r /ipfs/$CAR_ROOT_CID/examples/410.html | cut -d "/" -f3)
	custom410 := fixture.MustGetNode("examples", "410.html")
	// CUSTOM_4XX_CID=$(ipfs resolve -r /ipfs/$CAR_ROOT_CID/examples/451.html | cut -d "/" -f3)
	custom451 := fixture.MustGetNode("examples", "451.html")

	tests := SugarTests{
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
				URL("http://%s/redirect-one", redirectDirHostname),
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
				URL("http://%s/301-redirect-one", redirectDirHostname),
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
				URL("http://%s/302-redirect-two", redirectDirHostname),
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
				URL("http://%s/200-index", redirectDirHostname),
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
				URL("http://%s/posts/2022/01/01/hello-world", redirectDirHostname),
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
				URL("http://%s/splat/one.html", redirectDirHostname),
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
				URL("http://%s/not-found/has-no-redirects-entry", redirectDirHostname),
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
				URL("http://%s/gone/has-no-redirects-entry", redirectDirHostname),
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
				URL("http://%s/unavail/has-no-redirects-entry", redirectDirHostname),
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
				URL("http://%s/catch-all", redirectDirHostname),
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
	}

	// # Invalid file, containing forced redirect
	// INVALID_REDIRECTS_DIR_CID=$(ipfs resolve -r /ipfs/$CAR_ROOT_CID/forced | cut -d "/" -f3)
	invalidRedirectsDirCID := fixture.MustGetNode("forced").Base32Cid()
	// INVALID_REDIRECTS_DIR_HOSTNAME="${INVALID_REDIRECTS_DIR_CID}.ipfs.localhost:$GWAY_PORT"
	invalidDirHostname := fmt.Sprintf("%s.ipfs.localhost:8080", invalidRedirectsDirCID)
	// TOO_LARGE_REDIRECTS_DIR_CID=$(ipfs resolve -r /ipfs/$CAR_ROOT_CID/too-large | cut -d "/" -f3)
	tooLargeRedirectsDirCID := fixture.MustGetNode("too-large").Base32Cid()
	// TOO_LARGE_REDIRECTS_DIR_HOSTNAME="${TOO_LARGE_REDIRECTS_DIR_CID}.ipfs.localhost:$GWAY_PORT"
	tooLargeDirHostname := fmt.Sprintf("%s.ipfs.localhost:8080", tooLargeRedirectsDirCID)

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
				URL("http://%s/not-found", invalidDirHostname),
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
				URL("http://%s/not-found", tooLargeDirHostname),
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

	if specs.SubdomainGateway.IsEnabled() {
		Run(t, tests.Build())
	} else {
		t.Skip("subdomain gateway disabled")
	}
}

// TODO: dnslink tests