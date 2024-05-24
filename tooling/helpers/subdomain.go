package helpers

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/test"
)

/**
 * UnwrapSubdomainTests takes a list of tests and returns a (larger) list of tests
 * that will run on the subdomain gateway.
 */
func UnwrapSubdomainTests(t *testing.T, tests test.SugarTests) test.SugarTests {
	t.Helper()

	var out test.SugarTests
	for _, test := range tests {
		out = append(out, unwrapSubdomainTest(t, test)...)
	}
	return out
}

func unwrapSubdomainTest(t *testing.T, unwraped test.SugarTest) test.SugarTests {
	t.Helper()

	var baseURL, rawURL string
	req := unwraped.Request
	expected := unwraped.Response
	host := req.GetHeader("Host")
	if host != "" {
		// when custom Host header and Path are present we skip legacy magic
		// and use them as-is
		u, err := url.Parse(test.GatewayURL)
		if err != nil {
			panic("failed to parse GatewayURL")
		}
		// rawURL is gateway-url + Path
		u.Path = unwraped.Request.Path_
		unwraped.Request.Path_ = ""
		rawURL = u.String()
		// baseURL is rawURL with hostname from Host header
		u.Host = host
		baseURL = u.String()
		unwraped.Request.URL_ = baseURL
	} else {
		// Legacy flow based on URL instead of Host header
		baseURL := unwraped.Request.GetURL()

		u, err := url.Parse(baseURL)
		if err != nil {
			t.Fatal(err)
		}

		// change the low level HTTP endpoint to one defined via --gateway-url
		// to allow testing Host-based logic against arbitrary gateway URL (useful on CI)
		u.Host = test.GatewayHost

		rawURL = u.String()
	}

	// TODO: we want to refactor this magic into explicit Proxy test suite.
	// Having this magic here silently modifies headers such as Host, and if a
	// test fails, it is difficult to grasp how much really  is broken, because
	// number of errors is always multiplied x3. We should have standalone
	// proxy test for subdomain gateway and dnslink (simple GET should be
	// enough) and remove need for UnwrapSubdomainTests.

	return test.SugarTests{
		{
			Name: fmt.Sprintf("%s (direct HTTP)", unwraped.Name),
			Hint: fmt.Sprintf("%s\n%s", unwraped.Hint, "direct HTTP request (hostname in URL, raw IP in Host header)"),
			Request: req.
				URL(rawURL).
				Headers(
					test.Header("Host", host),
				),
			Response: expected,
		},
		{
			Name: fmt.Sprintf("%s (HTTP proxy)", unwraped.Name),
			Hint: fmt.Sprintf("%s\n%s", unwraped.Hint, "HTTP proxy (hostname is passed via URL)"),
			Request: req.
				URL(baseURL).
				Proxy(test.GatewayURL),
			Response: expected,
		},
		{
			Name: fmt.Sprintf("%s (HTTP proxy tunneling via CONNECT)", unwraped.Name),
			Hint: fmt.Sprintf("%s\n%s", unwraped.Hint, `HTTP proxy
				In HTTP/1.x, the pseudo-method CONNECT,
				can be used to convert an HTTP connection into a tunnel to a remote host
				https://tools.ietf.org/html/rfc7231#section-4.3.6
			`),
			Request: req.
				URL(baseURL).
				Proxy(test.GatewayURL).
				WithProxyTunnel().
				Headers(
					test.Header("Host", host),
				),
			Response: expected,
		},
	}
}
