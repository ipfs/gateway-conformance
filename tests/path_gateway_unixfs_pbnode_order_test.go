package tests

import (
	"encoding/hex"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

// PBNode field ordering fixtures from IPIP-550
// (https://github.com/ipfs/specs/pull/550): the same single-entry
// UnixFS Directory and HAMTShard encoded twice, once with the Data
// field before Links (streaming-friendly order written by the
// unixfs-v1-2026 profile) and once with Links before Data (legacy
// order written by unixfs-v1-2025 and earlier). Both orders are valid
// dag-pb and gateways must resolve content through either. Blocks are
// provisioned from fixtures/path_gateway_unixfs/pbnode-field-orders.car
// and asserted byte-exact against the hex in the IPIP fixtures table.
var (
	pbnodeOrderLeafCID  = "bafkreicysg23kiwv34eg2d7qweipxwosdo2py4ldv42nbauguluen5v6am"
	pbnodeOrderLeafData = mustHexDecode("68656c6c6f0a") // "hello\n"

	pbnodeOrderRoots = []struct {
		Name string
		Cid  string
		Data []byte
	}{
		{
			Name: "Directory with Data field before Links",
			Cid:  "bafybeigqvyloizmfcdy6scaxnyltftzptaruqa3hnnplfzsbf4sqteiwlm",
			Data: mustHexDecode("0a02080112330a24015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03120968656c6c6f2e7478741806"),
		},
		{
			Name: "Directory with Links before Data field",
			Cid:  "bafybeigdcg7pksx2zk5336vrfsktjodlr4rbfz37qr3koc5xboxe5ekv24",
			Data: mustHexDecode("12330a24015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03120968656c6c6f2e74787418060a020801"),
		},
		{
			Name: "HAMTShard with Data field before Links",
			Cid:  "bafybeicwgy2rlqmqqu3yy2tqvm2wbgdvy3snu4sbbv4wqpvpnoplpzxz74",
			Data: mustHexDecode("0a250805121c80000000000000000000000000000000000000000000000000000000282230800212350a24015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03120b444668656c6c6f2e7478741806"),
		},
		{
			Name: "HAMTShard with Links before Data field",
			Cid:  "bafybeicjwkfslu7gwyywffvqgse5kiibojtktxcdqhgv7ldj5fjdacuceq",
			Data: mustHexDecode("12350a24015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03120b444668656c6c6f2e74787418060a250805121c800000000000000000000000000000000000000000000000000000002822308002"),
		},
	}
)

func mustHexDecode(s string) []byte {
	data, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

// TestUnixFSPBNodeFieldOrder asserts that path resolution works through
// UnixFS directories and HAMT shards regardless of PBNode field order.
func TestUnixFSPBNodeFieldOrder(t *testing.T) {
	tooling.LogTestGroup(t, GroupUnixFS)

	tests := SugarTests{}
	for _, root := range pbnodeOrderRoots {
		tests = append(tests, SugarTest{
			Name: "GET file from " + root.Name,
			Hint: "both PBNode field orders decode to the same logical node, so pathing must work through either encoding",
			Request: Request().
				Path("/ipfs/{{cid}}/hello.txt", root.Cid),
			Response: Expect().
				Status(200).
				Body(pbnodeOrderLeafData),
		})
	}

	RunWithSpecs(t, tests, specs.PathGatewayUnixFS)
}

// TestTrustlessRawPBNodeFieldOrder asserts that raw block responses
// return verbatim bytes for both PBNode field orders, without
// re-encoding to a preferred order (which would change the CID).
func TestTrustlessRawPBNodeFieldOrder(t *testing.T) {
	tooling.LogTestGroup(t, GroupBlockCar)

	tests := SugarTests{
		{
			Name: "GET raw leaf block referenced by both field orders",
			Request: Request().
				Path("/ipfs/{{cid}}", pbnodeOrderLeafCID).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
				),
			Response: Expect().
				Status(200).
				Body(pbnodeOrderLeafData),
		},
	}
	for _, root := range pbnodeOrderRoots {
		tests = append(tests, SugarTest{
			Name: "GET raw block of " + root.Name,
			Hint: "raw block response bytes must match the stored block exactly, preserving PBNode field order",
			Request: Request().
				Path("/ipfs/{{cid}}", root.Cid).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
				),
			Response: Expect().
				Status(200).
				Body(root.Data),
		})
	}

	RunWithSpecs(t, tests, specs.TrustlessGatewayRaw)
}
