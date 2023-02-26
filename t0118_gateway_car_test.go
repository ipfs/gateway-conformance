package main

import (
	"testing"

	"github.com/ipfs/gateway-conformance/test"
)

func TestGatewayCar(t *testing.T) {
	tests := map[string]test.Test{}

	test.Run(t, tests)
}
