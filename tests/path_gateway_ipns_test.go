package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

var (
	// See fixtures/ipns_records/README.md for information
	ipnsV1                = "k51qzi5uqu5dm4tm0wt8srkg9h9suud4wuiwjimndrkydqm81cqtlb5ak6p7ku"
	ipnsV1V2BrokenValueV1 = "k51qzi5uqu5dlmit2tuwdvnx4sbnyqgmvbxftl0eo3f33wwtb9gr7yozae9kpw"
	ipnsV1V2BrokenSigV2   = "k51qzi5uqu5diamp7qnnvs1p1gzmku3eijkeijs3418j23j077zrkok63xdm8c"

	ipnsV1V2BrokenSigV1     = "k51qzi5uqu5dilgf7gorsh9vcqqq4myo6jd4zmqkuy9pxyxi5fua3uf7axph4y"
	bodyIPNSV1V2BrokenSigV1 = []byte("v1+v2 with broken signature v1")

	ipnsV1V2     = "k51qzi5uqu5dlkw8pxuw9qmqayfdeh4kfebhmreauqdc6a7c3y7d5i9fi8mk9w"
	bodyIPNSV1V2 = []byte("v1+v2 record")
	cidIPNSV1V2  = "bafkqaddwgevxmmraojswg33smq"

	ipnsV2     = "k51qzi5uqu5dit2ku9mutlfgwyz8u730on38kd10m97m36bjt66my99hb6103f"
	bodyIPNSV2 = []byte("v2-only record")
	cidIPNSV2  = "bafkqadtwgiww63tmpeqhezldn5zgi"
)

func TestGatewayIPNSPath(t *testing.T) {
	tests := SugarTests{
		{
			Name: "GET an IPNS Path (V1) from the gateway fails with 5xx",
			Request: Request().
				Path("/ipns/{{name}}", ipnsV1),
			Response: Expect().
				StatusRange(500, 599),
		},
		{
			Name: "GET an IPNS Path (V1+V2) with broken ValueV1 from the gateway fails with 5xx",
			Request: Request().
				Path("/ipns/{{name}}", ipnsV1V2BrokenValueV1),
			Response: Expect().
				StatusRange(500, 599),
		},
		{
			Name: "GET an IPNS Path (V1+V2) with broken SignatureV1, but valid SignatureV2 succeeds",
			Request: Request().
				Path("/ipns/{{name}}", ipnsV1V2BrokenSigV1),
			Response: Expect().
				Status(200).
				Body(bodyIPNSV1V2BrokenSigV1),
		},
		{
			Name: "GET an IPNS Path (V1+V2) from the gateway",
			Request: Request().
				Path("/ipns/{{name}}", ipnsV1V2),
			Response: Expect().
				Body(bodyIPNSV1V2),
		},
		{
			Name: "GET an IPNS Path (V2) from the gateway",
			Request: Request().
				Path("/ipns/{{name}}", ipnsV2),
			Response: Expect().
				Body(bodyIPNSV2),
		},
		{
			Name: "GET an IPNS Path (V1+V2) with broken SignatureV2 from the gateway fails with 5xx",
			Request: Request().
				Path("/ipns/{{name}}", ipnsV1V2BrokenSigV2),
			Response: Expect().
				StatusRange(500, 599),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayIPNS)
}

func TestRedirectCanonicalIPNS(t *testing.T) {
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
