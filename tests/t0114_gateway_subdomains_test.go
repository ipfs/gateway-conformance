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

	// skipped: # IP remains old school path-based gateway

	// 'localhost' hostname is used for subdomains, and should not return
	//  payload directly, but redirect to URL with proper origin isolation
	tests = append(tests, testLocalhostGatewayResponseShouldContain(t,
		"request for localhost/ipfs/{cid} redirects to subdomain",
		// TODO: this works with the `/` suffix only. Why?
		fmt.Sprintf("http://localhost/ipfs/%s/", CIDv1),
		Expect().
			Status(301).
			Headers(
				Header("Location").
					Hint("request for localhost/ipfs/{CIDv1} returns Location HTTP header for subdomain redirect in browsers").
					Contains("http://%s.ipfs.localhost", CIDv1),
			).Response(),
	)...)

	tests = append(tests, testLocalhostGatewayResponseShouldContain(t,
		"request for localhost/ipfs/{directory} redirects to subdomain",
		// TODO: this works with the `/` suffix only. Why?
		fmt.Sprintf("http://localhost/ipfs/%s/", DirCID),
		Expect().
			Status(301).
			Headers(
				Header("Location").
					Hint("request for localhost/ipfs/{CIDv1} returns Location HTTP header for subdomain redirect in browsers").
					Contains("http://%s.ipfs.localhost", DirCID),
			).Response(),
	)...)

	tests = append(tests, []CTest{
		{

			Name: "request for {gateway}/ipfs/{CIDv1} returns HTTP 301 Moved Permanently (sugar)",
			Request: Request().
				URL("http://example.com/ipfs/%s", CIDv1).
				DoNotFollowRedirects().
				Request(),
			Response: Expect().
				Status(301).
				Headers(
					Header("Location").
						// TODO: this works only because we use example.com in our tests.
						// It should be:
						// Contains("%s://%s.ipfs.%s", SubdomainGatewayScheme, CIDv1, SubdomainGatewayHost)
						// I am trying to avoid this syntax.
						// The other option is to force the tested gateway to use example.com.
						Contains("http://%s.ipfs.example.com", CIDv1),
				).
				Response(),
		},
		{
			Name: "request for {cid}.ipfs.localhost/api returns data if present on the content root (sugar)",
			Request: Request().
				URL("http://%s.ipfs.example.com/api/file.txt", DirCID).
				Request(),
			Response: Expect().
				Status(200).
				Body("I am a txt file\n").
				Response(),
		},
	}...)

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
	hostName := u.Hostname()

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
					Header("Host", hostName),
				).
				Request(),
			Response: expected,
		},
		{
			Name: fmt.Sprintf("%s (HTTP proxy)", label),
			Hint: "HTTP proxy (hostname is passed via URL)",
			Request: Request().
				URL(myUrl).
				DoNotFollowRedirects().
				Request(),
			Response: expected,
		},
	}
}
