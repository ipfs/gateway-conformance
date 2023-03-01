package main

import (
	"testing"

	"github.com/ipfs/gateway-conformance/test"
)

func TestGatewayTar(t *testing.T) {
	tests := []test.CTest{}

	test.Run(t, tests)
}
