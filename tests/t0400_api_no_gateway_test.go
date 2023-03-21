package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/test"
)

func TestApiNoGateway(t *testing.T) {
	tests := []test.CTest{}

	test.Run(t, tests)
}
