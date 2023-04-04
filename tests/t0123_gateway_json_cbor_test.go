package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewayJsonCbor(t *testing.T) {
	tests := test.SugarTests{}

	test.Run(t, tests)
}
