package tests

import (
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
		// TODO: path gw: redirect dir listing to URL with trailing slash
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

	// TODO: fix the subdomain gateway test
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
			Request: Request().
				URL("%s://%s.%s:%s/", u.Scheme, dir.Cid(), u.Host, u.Port),
			Response: Expect().
				Status(200).
				Body(
					And(
						Contains("Index of"),
						Contains("<a href=\"/\">..</a>"),
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
