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
	GroupTrustlessGateway = "Trustless Gateway"
	GroupPathGateway      = "Path Gateway"
	GroupSubdomainGateway = "Subdomain Gateway"
	GroupCORS             = "CORS"
	GroupIPNS             = "IPNS"
)
