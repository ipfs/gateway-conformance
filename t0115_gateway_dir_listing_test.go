package main

import (
	"testing"

	"github.com/ipfs/gateway-conformance/test"
)

func TestGatewayDirListing(t *testing.T) {
	tests := map[string]test.Test{}

	test.Run(t, tests)
}
