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
	dnsLinkB58 := dnsLinks.MustGet("dnslink-over-ipns")
	dnsLinkCIDv1 := dnsLinks.MustGet("dnslink-over-ipns-cidv1")

	// same content the Ed25519 IPNS record from subdomain_gateway points at
	fixture := car.MustOpenUnixfsCar("subdomain_gateway/fixtures.car")
	payload := string(fixture.MustGetRawData("hello-CIDv1"))

	tests := SugarTests{
		{
			Name: "GET for DNSLink with dnslink=/ipns/{peer-id} returns expected payload",
			Hint: `
			When a DNSLink TXT record points to /ipns/<peer-id> (base58btc),
			the gateway must resolve the IPNS name and then serve the content
			it points to.
			`,
			Request: Request().
				Header("Host", dnsLinkB58).
				Path("/"),
			Response: Expect().
				Status(200).
				Body(payload),
		},
		{
			Name: "GET for DNSLink with dnslink=/ipns/{cidv1-libp2p-key} returns expected payload",
			Hint: `
			When a DNSLink TXT record points to /ipns/<cidv1-libp2p-key> (base36),
			the gateway must resolve the IPNS name and then serve the content
			it points to.
			`,
			Request: Request().
				Header("Host", dnsLinkCIDv1).
				Path("/"),
			Response: Expect().
				Status(200).
				Body(payload),
		},
	}

	RunWithSpecs(t, tests, specs.DNSLinkGateway)
}
