package tests

import (
	"net/url"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling"
	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/dnslink"
	"github.com/ipfs/gateway-conformance/tooling/helpers"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
	. "github.com/ipfs/gateway-conformance/tooling/tmpl"
)

func TestRedirectsFileSupport(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("redirects_file/redirects.car")
	redirectDir := fixture.MustGetNode("examples")
	redirectDirCID := redirectDir.Base32Cid()

	custom404 := fixture.MustGetNode("examples", "404.html")
	custom410 := fixture.MustGetNode("examples", "410.html")
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

		redirectDirBaseURL := Fmt("{{scheme}}://{{cid}}.ipfs.{{host}}", u.Scheme, redirectDirCID, u.Host)

		tests = append(tests, SugarTests{
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/redirect-one redirects with default of 301, per _redirects file",
				Request: Request().
					Header("Host", u.Host).
					URL("{{url}}/redirect-one", redirectDirBaseURL),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("/one.html"),
					),
			},
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/301-redirect-one redirects with 301, per _redirects file",
				Request: Request().
					URL("{{url}}/301-redirect-one", redirectDirBaseURL),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("/one.html"),
					),
			},
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/302-redirect-two redirects with 302, per _redirects file",
				Request: Request().
					URL("{{url}}/302-redirect-two", redirectDirBaseURL),
				Response: Expect().
					Status(302).
					Headers(
						Header("Location").Equals("/two.html"),
					),
			},
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/200-index returns 200, per _redirects file",
				Request: Request().
					URL("{{url}}/200-index", redirectDirBaseURL),
				Response: Expect().
					Status(200).
					Body(Contains("my index")),
			},
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/posts/:year/:month/:day/:title redirects with 301 and placeholders, per _redirects file",
				Request: Request().
					URL("{{url}}/posts/2022/01/01/hello-world", redirectDirBaseURL),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("/articles/2022/01/01/hello-world"),
					),
			},
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/splat/one.html redirects with 301 and splat placeholder, per _redirects file",
				Request: Request().
					URL("{{url}}/splat/one.html", redirectDirBaseURL),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("/redirected-splat/one.html"),
					),
			},
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/not-found/has-no-redirects-entry returns custom 404, per _redirects file",
				Request: Request().
					URL("{{url}}/not-found/has-no-redirects-entry", redirectDirBaseURL),
				Response: Expect().
					Status(404).
					Headers(
						Header("Cache-Control").Equals("public, max-age=29030400, immutable"),
						Header("Etag").Equals(`"{{etag}}"`, custom404.Cid().String()),
					).
					Body(Contains(custom404.ReadFile())),
			},
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/gone/has-no-redirects-entry returns custom 410, per _redirects file",
				Request: Request().
					URL("{{url}}/gone/has-no-redirects-entry", redirectDirBaseURL),
				Response: Expect().
					Status(410).
					Headers(
						Header("Cache-Control").Equals("public, max-age=29030400, immutable"),
						Header("Etag").Equals(`"{{etag}}"`, custom410.Cid().String()),
					).
					Body(Contains(custom410.ReadFile())),
			},
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/unavail/has-no-redirects-entry returns custom 451, per _redirects file",
				Request: Request().
					URL("{{url}}/unavail/has-no-redirects-entry", redirectDirBaseURL),
				Response: Expect().
					Status(451).
					Headers(
						Header("Cache-Control").Equals("public, max-age=29030400, immutable"),
						Header("Etag").Equals(`"{{etag}}"`, custom451.Cid().String()),
					).
					Body(Contains(custom451.ReadFile())),
			},
			{
				Name: "request for $REDIRECTS_DIR_HOSTNAME/catch-all returns 200, per _redirects file",
				Request: Request().
					URL("{{url}}/catch-all", redirectDirBaseURL),
				Response: Expect().
					Status(200).
					Body(Contains("my index")),
			},
		}...)

		// # Invalid file, containing forced redirect
		invalidRedirectsDirCID := fixture.MustGetNode("forced").Base32Cid()
		invalidDirBaseURL := Fmt("{{scheme}}://{{cid}}.ipfs.{{host}}", u.Scheme, invalidRedirectsDirCID, u.Host)

		tooLargeRedirectsDirCID := fixture.MustGetNode("too-large").Base32Cid()
		tooLargeDirBaseURL := Fmt("{{scheme}}://{{cid}}.ipfs.{{host}}", u.Scheme, tooLargeRedirectsDirCID, u.Host)

		tests = append(tests, SugarTests{
			{
				Name: "invalid file: request for $INVALID_REDIRECTS_DIR_HOSTNAME/not-found returns error about invalid redirects file",
				Hint: `if accessing a path that doesn't exist, read _redirects and fail parsing, and return error`,
				Request: Request().
					URL("{{url}}/not-found", invalidDirBaseURL),
				Response: Expect().
					Status(500).
					Body(
						And(
							Contains("could not parse _redirects:"),
							Contains(`forced redirects (or "shadowing") are not supported`),
						),
					),
			},
			{
				Name: "invalid file: request for $TOO_LARGE_REDIRECTS_DIR_HOSTNAME/not-found returns error about too large redirects file",
				Hint: `if accessing a path that doesn't exist and _redirects file is too large, return error`,
				Request: Request().
					URL("{{url}}/not-found", tooLargeDirBaseURL),
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

		// # With CRLF line terminator
		newlineRedirectsDirCID := fixture.MustGetNode("newlines").Base32Cid()
		newlineBaseURL := Fmt("{{scheme}}://{{cid}}.ipfs.{{host}}", u.Scheme, newlineRedirectsDirCID, u.Host)

		// # Good codes
		goodRedirectDirCID := fixture.MustGetNode("good-codes").Base32Cid()
		goodRedirectDirBaseURL := Fmt("{{scheme}}://{{cid}}.ipfs.{{host}}", u.Scheme, goodRedirectDirCID, u.Host)

		// # Bad codes
		badRedirectDirCID := fixture.MustGetNode("bad-codes").Base32Cid()
		badRedirectDirBaseURL := Fmt("{{scheme}}://{{cid}}.ipfs.{{host}}", u.Scheme, badRedirectDirCID, u.Host)

		tests = append(tests, SugarTests{
			{
				Name: "newline: request for $NEWLINE_REDIRECTS_DIR_HOSTNAME/redirect-one redirects with default of 301, per _redirects file",
				Request: Request().
					URL("{{url}}/redirect-one", newlineBaseURL),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("/one.html"),
					),
			},
			{
				Name: "good codes: request for $GOOD_REDIRECTS_DIR_HOSTNAME/redirect-one redirects with default of 301, per _redirects file",
				Request: Request().
					URL("{{url}}/a301", goodRedirectDirBaseURL),
				Response: Expect().
					Status(301).
					Headers(
						Header("Location").Equals("/b301"),
					),
			},
			{
				Name: "bad codes: request for $BAD_REDIRECTS_DIR_HOSTNAME/found.html doesn't return error about bad code",
				Request: Request().
					URL("{{url}}/found.html", badRedirectDirBaseURL),
				Response: Expect().
					Status(200).
					Body(
						And(
							Contains("my found"),
							Not(Contains("unsupported redirect status")),
						),
					),
			},
		}...)
	}

	RunWithSpecs(t, helpers.UnwrapSubdomainTests(t, tests), specs.SubdomainGatewayIPFS, specs.RedirectsFile)
}

func TestRedirectsFileSupportWithDNSLink(t *testing.T) {
	tooling.LogTestGroup(t, GroupDNSLink)
	dnsLinks := dnslink.MustOpenDNSLink("redirects_file/dnslink.yml")
	dnsLink := dnsLinks.MustGet("custom-dnslink")

	gatewayURL := SubdomainGatewayURL
	u, err := url.Parse(gatewayURL)
	if err != nil {
		t.Fatal(err)
	}

	dnsLinkBaseUrl := Fmt("{{scheme}}://{{dnslink}}.{{host}}", u.Scheme, dnsLink, u.Host)

	tests := SugarTests{
		{
			Name: "request for $DNSLINK_FQDN/redirect-one redirects with default of 301, per _redirects file",
			Request: Request().
				URL("{{url}}/redirect-one", dnsLinkBaseUrl),
			Response: Expect().
				Status(301).
				Headers(
					Header("Location", "/one.html"),
				),
		},
		{
			Name: "request for $DNSLINK_FQDN/en/has-no-redirects-entry returns custom 404, per _redirects file",
			Hint: `ensure custom 404 works and has the same cache headers as regular /ipns/ paths`,
			Request: Request().
				URL("{{url}}/not-found/has-no-redirects-entry", dnsLinkBaseUrl),
			Response: Expect().
				Status(404).
				Headers(
					Header("Etag", `"Qmd9GD7Bauh6N2ZLfNnYS3b7QVAijbud83b8GE8LPMNBBP"`),
					Header("Cache-Control").Not().Contains("public, max-age=29030400, immutable"),
					Header("Cache-Control").Not().Contains("immutable"),
					Header("Date").Exists(),
				).
				Body(
					Contains("my 404"),
				),
		},
	}

	RunWithSpecs(t, helpers.UnwrapSubdomainTests(t, tests), specs.DNSLinkGateway, specs.RedirectsFile)
}
