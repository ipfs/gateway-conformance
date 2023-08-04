package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling"
	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/ipns"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewayIPNSRecord(t *testing.T) {
	tooling.LogTestGroup(t, GroupTrustlessGateway)

	fixture := car.MustOpenUnixfsCar("ipns_records/fixtures.car")
	file := fixture.MustGetRoot()
	fileCID := file.Cid()

	ipns := MustOpenIPNSRecordWithKey("ipns_records/k51qzi5uqu5dh71qgwangrt6r0nd4094i88nsady6qgd1dhjcyfsaqmpp143ab.ipns-record")
	ipnsName := ipns.Key()

	tests := SugarTests{
		{
			Name: "GET an IPNS path from the gateway",
			Request: Request().
				Path("/ipns/{{name}}", ipnsName),
			Response: Expect().
				Body(file.RawData()),
		},
		{
			Name: "GET IPNS Record with format=ipns-record has expected HTTP headers and valid key",
			Request: Request().
				Path("/ipns/{{name}}", ipnsName).
				Query("format", "ipns-record"),
			Response: Expect().
				Headers(
					Header("Content-Disposition").Contains("attachment;"),
					Header("Content-Type").Contains("application/vnd.ipfs.ipns-record"),
					Header("Cache-Control").Contains("public, max-age=3155760000"),
				).
				Body(
					IsIPNSRecord(ipnsName).
						IsValid().
						PointsTo("/ipfs/{{cid}}", fileCID.String()),
				),
		},
		{
			Name: "GET IPNS Record with 'Accept: application/vnd.ipfs.ipns-record' has expected HTTP headers and valid key",
			Request: Request().
				Path("/ipns/{{name}}", ipnsName).
				Header("Accept", "application/vnd.ipfs.ipns-record"),
			Response: Expect().
				Headers(
					Header("Content-Disposition").Contains("attachment;"),
					Header("Content-Type").Contains("application/vnd.ipfs.ipns-record"),
					Header("Cache-Control").Contains("public, max-age=3155760000"),
				).
				Body(
					IsIPNSRecord(ipnsName).
						IsValid().
						PointsTo("/ipfs/{{cid}}", fileCID.String()),
				),
		},
		{
			Name: "GET IPNS Record with explicit ?filename= succeeds with modified Content-Disposition header",
			Request: Request().
				Path("/ipns/{{name}}", ipnsName).
				Query("format", "ipns-record").
				Query("filename", "testтест.ipns-record"),
			Response: Expect().
				Headers(
					Header("Content-Disposition").
						Contains(`attachment; filename="test____.ipns-record"; filename*=UTF-8''test%D1%82%D0%B5%D1%81%D1%82.ipns-record`),
				),
		},
	}

	RunWithSpecs(t, tests, specs.TrustlessGatewayIPNS)
}
