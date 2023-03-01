package main

import (
	"testing"

	"github.com/ipfs/gateway-conformance/test"
)

func TestApiNoGateway(t *testing.T) {
	tests := []test.CTest{}

	test.Run(t, tests)
}
