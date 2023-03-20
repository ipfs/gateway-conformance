package tests

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewaySubdomains(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0114-gateway_subdomains.car")

	DirCID := fixture.MustGetCid("testdirlisting")
	CIDv1 := fixture.MustGetCid("hello-CIDv1")
	CIDv0 := fixture.MustGetCid("hello-CIDv0")
	CIDv0to1 := fixture.MustGetCid("hello-CIDv0to1")
	CIDv1_TOO_LONG := fixture.MustGetCid("hello-CIDv1_TOO_LONG")

	fmt.Println("DIR_CID:", DirCID)
	fmt.Println("CIDv1:", CIDv1, "CIDv0:", CIDv0)
	fmt.Println("CIDv0to1:", CIDv0to1, "CIDv1_TOO_LONG:", CIDv1_TOO_LONG)

	tests := []CTest{}

	gatewayURLs := []string{
		SubdomainGatewayUrl,
		SubdomainLocalhostGatewayUrl,
	}

	for _, gatewayURL := range gatewayURLs {
		u, err := url.Parse(gatewayURL)
		if err != nil {
			t.Fatal(err)
		}

		// skipped: # IP remains old school path-based gateway

		// 'localhost' hostname is used for subdomains, and should not return
		//  payload directly, but redirect to URL with proper origin isolation
		tests = append(tests, testLocalhostGatewayResponseShouldContain(t,
			"request for localhost/ipfs/{cid} redirects to subdomain",
			fmt.Sprintf("%s/ipfs/%s/", gatewayURL, CIDv1),
			Expect().
				Status(301).
				Headers(
					Header("Location").
						Hint("request for localhost/ipfs/{CIDv1} returns Location HTTP header for subdomain redirect in browsers").
						Contains("%s://%s.ipfs.%s/", u.Scheme, CIDv1, u.Host),
				).Response(),
		)...)

	}

	// tests = append(tests, testLocalhostGatewayResponseShouldContain(t,
	// 	"request for localhost/ipfs/{directory} redirects to subdomain",
	// 	// TODO: this works with the `/` suffix only. Why?
	// 	fmt.Sprintf("http://localhost/ipfs/%s/", DirCID),
	// 	Expect().
	// 		Status(301).
	// 		Headers(
	// 			Header("Location").
	// 				Hint("request for localhost/ipfs/{CIDv1} returns Location HTTP header for subdomain redirect in browsers").
	// 				Contains("http://%s.ipfs.localhost", DirCID),
	// 		).Response(),
	// )...)

	// tests = append(tests, []CTest{
	// 	{

	// 		Name: "request for {gateway}/ipfs/{CIDv1} returns HTTP 301 Moved Permanently",
	// 		Request: Request().
	// 			URL("%s/ipfs/%s", SubdomainGatewayUrl, CIDv1).
	// 			DoNotFollowRedirects().
	// 			Request(),
	// 		Response: Expect().
	// 			Status(301).
	// 			Headers(
	// 				Header("Location").
	// 					Contains("%s://%s.ipfs.%s", SubdomainGatewayScheme, CIDv1, SubdomainGatewayHost),
	// 			).
	// 			Response(),
	// 	},
	// 	{
	// 		Name: "request for {cid}.ipfs.example.com/api returns data if present on the content root",
	// 		Request: Request().
	// 			URL("%s://%s.ipfs.%s/api/file.txt", SubdomainGatewayScheme, DirCID, SubdomainGatewayHost).
	// 			Request(),
	// 		Response: Expect().
	// 			Status(200).
	// 			Body("I am a txt file\n").
	// 			Response(),
	// 	},
	// }...)

	if SubdomainGateway.IsEnabled() {
		Run(t, tests)
	}
}

func testLocalhostGatewayResponseShouldContain(t *testing.T, label string, myUrl string, expected CResponse) []CTest {
	t.Helper()

	u, err := url.Parse(myUrl)
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
			Hint: "regular HTTP request (hostname in Host header, raw IP in URL)",
			Request: Request().
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
			Hint: "HTTP proxy (hostname is passed via URL)",
			Request: Request().
				URL(myUrl).
				Proxy(GatewayUrl).
				DoNotFollowRedirects().
				Request(),
			Response: expected,
		},
		// TODO: port the proxy1.0 as well (use http 1) but also http2?
		{
			Name: fmt.Sprintf("%s (HTTP proxy tunneling via CONNECT)", label),
			Hint: `HTTP proxy
				In HTTP/1.x, the pseudo-method CONNECT,
				can be used to convert an HTTP connection into a tunnel to a remote host
				https://tools.ietf.org/html/rfc7231#section-4.3.6
			`,
			Request: Request().
				URL(myUrl).
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
