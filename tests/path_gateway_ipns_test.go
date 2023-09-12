package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestRedirectCanonicalIPNS(t *testing.T) {
	tooling.LogTestGroup(t, GroupIPNS)

	tests := SugarTests{
		{
			Name: "GET for /ipns/{b58-multihash-of-ed25519-key} redirects to /ipns/{cidv1-libp2p-key-base36}",
			Request: Request().
				Path("/ipns/12D3KooWRBy97UB99e3J6hiPesre1MZeuNQvfan4gBziswrRJsNK/root2/"),
			Response: Expect().
				Status(302).
				Headers(
					Header("Location").Equals("/ipns/k51qzi5uqu5dlvj2baxnqndepeb86cbk3ng7n3i46uzyxzyqj2xjonzllnv0v8/root2/"),
				),
		},
		{
			Name: "GET for /ipns/{cidv0-like-b58-multihash-of-rsa-key} redirects to /ipns/{cidv1-libp2p-key-base36}",
			Request: Request().
				Path("/ipns/QmcJM7PRfkSbcM5cf1QugM5R37TLRKyJGgBEhXjLTB8uA2/root2/"),
			Response: Expect().
				Status(302).
				Headers(
					Header("Location").Equals("/ipns/k2k4r8ol4m8kkcqz509c1rcjwunebj02gcnm5excpx842u736nja8ger/root2/"),
				),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayIPNS)
}
