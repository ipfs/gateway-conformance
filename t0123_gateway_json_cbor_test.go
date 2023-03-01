package main

import (
	"testing"

	"github.com/ipfs/gateway-conformance/test"
)

func TestGatewayJsonCbor(t *testing.T) {
	tests := []test.CTest{}

	test.Run(t, tests)
}
