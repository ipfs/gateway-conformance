package main

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewayJsonCbor(t *testing.T) {
	tests := []test.CTest{}

	test.Run(t, tests)
}
