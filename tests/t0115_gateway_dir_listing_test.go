package tests

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/helpers"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	"github.com/ipfs/gateway-conformance/tooling/test"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewayDirListingOnPathGateway(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0115/fixtures.car")
	dir := fixture.MustGetRoot()
	// file := fixture.MustGetNode("ą", "ę", "file-źł.txt")

	tests := SugarTests{
		// test_expect_success "path gw: backlink on root CID should be hidden" '
		//   curl -sD - http://127.0.0.1:$GWAY_PORT/ipfs/${DIR_CID}/ > list_response &&
		//   test_should_contain "Index of" list_response &&
		//   test_should_not_contain "<a href=\"/ipfs/$DIR_CID/\">..</a>" list_response
		// '
		{
			// TODO: run the test, check the report and fix this test.
			Name: "path gw: backlink on root CID should be hidden",
			Hint: `
			this test is written for the workshop, it will fail by default.
			But we can use it to show the rough idea of how to write tests.
			`,
			Request: Request().
				Path("/ipfs/%s/", dir.Cid()),
			Response: Expect().
				Status(202).
				Body(
					And(
						Contains("Index of"),
						Not(Contains("<a href=\"/ipfs/%s/\">..</a>", dir.Cid())),
					),
				),
		},
		// test_expect_success "path gw: redirect dir listing to URL with trailing slash" '
		//   curl -sD - http://127.0.0.1:$GWAY_PORT/ipfs/${DIR_CID}/ą/ę > list_response &&
		//   test_should_contain "HTTP/1.1 301 Moved Permanently" list_response &&
		//   test_should_contain "Location: /ipfs/${DIR_CID}/%c4%85/%c4%99/" list_response
		// '
		// TODO: implement this test.
	}

	test.Run(t, tests)
}

func TestGatewayDirListingOnSubdomainGateway(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0115/fixtures.car")
	dir := fixture.MustGetRoot()
	// file := fixture.MustGetNode("ą", "ę", "file-źł.txt")

	gatewayURL := SubdomainLocalhostGatewayURL

	u, err := url.Parse(gatewayURL)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: figure out how to use these informations below:
	fmt.Printf("gatewayURL: %s, dirCid: %s, host: %s", gatewayURL, dir.Cid(), u.Host)

	tests := SugarTests{
		// test_expect_success "subdomain gw: backlink on root CID should be hidden" '
		//   curl -sD - --resolve $DIR_HOSTNAME:$GWAY_PORT:127.0.0.1 http://$DIR_HOSTNAME:$GWAY_PORT/ > list_response &&
		//   test_should_contain "Index of" list_response &&
		//   test_should_not_contain "<a href=\"/\">..</a>" list_response
		// '
		{
			Name: "subdomain gw: backlink on root CID should be hidden",
			Hint: `
			This test is using a custom configuration to resolve the hostname to an IP address.
			`,
			Request: Request(),
			// URL("%s://%s.%s/", u.Scheme, dir.Cid(), u.Host),
			Response: Expect().
				Status(200).
				Body(
					And(
						Contains("Index of"),
						Not(Contains("<a href=\"/\">..</a>")),
					),
				),
		},
	}

	test.RunIfSpecsAreEnabled(
		t,
		helpers.UnwrapSubdomainTests(t, tests),
		specs.SubdomainGateway,
	)
}

func TestGatewayDirListingOnDNSLinkGateway(t *testing.T) {
	// fixture := car.MustOpenUnixfsCar("t0115/fixtures.car")
	// dir := fixture.MustGetRoot()
	// file := fixture.MustGetNode("ą", "ę", "file-źł.txt")

	tests := SugarTests{}

	test.RunIfSpecsAreEnabled(
		t,
		tests,
		specs.DNSLinkResolver,
	)
}
