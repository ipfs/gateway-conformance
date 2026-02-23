package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling"
	"github.com/ipfs/gateway-conformance/tooling/car"
	"github.com/ipfs/gateway-conformance/tooling/dnslink"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestDNSLinkGatewayIPNS(t *testing.T) {
	tooling.LogTestGroup(t, GroupDNSLink)

	dnsLinks := dnslink.MustOpenDNSLink("dnslink_ipns/dnslink.yml")
	dnsLinkDomain := dnsLinks.MustGet("dnslink-over-ipns")

	// same content the Ed25519 IPNS record from subdomain_gateway points at
	fixture := car.MustOpenUnixfsCar("subdomain_gateway/fixtures.car")
	payload := string(fixture.MustGetRawData("hello-CIDv1"))

	tests := SugarTests{
		{
			Name: "GET for DNSLink domain with dnslink=/ipns/{key} returns expected payload",
			Hint: `
			When a DNSLink TXT record points to /ipns/<key> instead of /ipfs/<cid>,
			the gateway must first resolve the IPNS name and then serve the content
			it points to.
			`,
			Request: Request().
				Header("Host", dnsLinkDomain).
				Path("/"),
			Response: Expect().
				Status(200).
				Body(payload),
		},
	}

	RunWithSpecs(t, tests, specs.DNSLinkGateway)
}
