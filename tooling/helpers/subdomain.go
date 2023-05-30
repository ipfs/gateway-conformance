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

	baseURL := unwraped.Request.GetURL()
	req := unwraped.Request
	expected := unwraped.Response

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
	u.Host = test.GatewayHost
	rawURL := u.String()

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
