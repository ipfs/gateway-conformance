package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling"
)

func TestMetadata(t *testing.T) {
	tooling.LogVersion(t)
	tooling.LogJobURL(t)
	tooling.LogGatewayURL(t)
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
