package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling"
	"github.com/ipfs/gateway-conformance/tooling/test"
)

func logGatewayURL(t *testing.T) {
	tooling.LogMetadata(t, struct {
		GatewayURL          string `json:"gateway_url"`
		SubdomainGatewayURL string `json:"subdomain_gateway_url"`
	}{
		GatewayURL:          test.GatewayURL,
		SubdomainGatewayURL: test.SubdomainGatewayURL,
	})
}

func TestMetadata(t *testing.T) {
	tooling.LogVersion(t)
	tooling.LogJobURL(t)
	logGatewayURL(t)
}

const (
	GroupSubdomains = "Subdomains"
	GroupCORS       = "CORS"
	GroupIPNS       = "IPNS"
	GroupDNSLink    = "DNSLink"
	GroupJSONCbor   = "JSON-CBOR"
	GroupBlockCar   = "Block-CAR"
	GroupTar        = "Tar"
	GroupUnixFS     = "UnixFS"
)
