package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewayTar(t *testing.T) {
	tests := test.SugarTests{}

	test.Run(t, tests)
}
