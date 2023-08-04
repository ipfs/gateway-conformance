package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling"
)

func TestMetadata(t *testing.T) {
	tooling.LogVersion(t)
}

const (
	GroupTrustlessGateway = "Trustless Gateway"
	GroupPathGateway      = "Path Gateway"
	GroupSubdomainGateway = "Subdomain Gateway"
)

const (
	IPIP402 = "ipip-0402"
)
