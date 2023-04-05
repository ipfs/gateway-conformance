package tests

import (
	"fmt"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/test"
	. "github.com/ipfs/gateway-conformance/tooling/test"
	mb "github.com/multiformats/go-multibase"
)

func TestRedirectsFileSupport(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0109-redirects.car")

	// root := fixture.MustGetNode()
	redirectDir := fixture.MustGetNode("examples")

	redirectDirCID, err := mb.Encode(mb.Base32, redirectDir.Cid().Bytes())
	if err != nil {
		t.Fatal(err)
	}

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

	test.Run(t, tests.Build())
}
