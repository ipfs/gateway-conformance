package tests

import (
	"testing"

	. "github.com/ipfs/gateway-conformance/tooling/ipns"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewayIPNSRecord(t *testing.T) {
	tests := SugarTests{
		{
			Name: "GET IPNS Record (V1+V2) with format=ipns-record has expected HTTP headers and valid key",
			Request: Request().
				Path("/ipns/{{name}}", ipnsV1V2).
				Query("format", "ipns-record"),
			Response: Expect().
				Headers(
					Header("Content-Disposition").Contains("attachment;"),
					Header("Content-Type").Contains("application/vnd.ipfs.ipns-record"),
					Header("Cache-Control").Contains("public, max-age=1800"),
				).
				Body(
					IsIPNSRecord(ipnsV1V2).
						IsValid().
						PointsTo("/ipfs/{{cid}}", cidIPNSV1V2),
				),
		},
		{
			Name: "GET IPNS Record (V2) with format=ipns-record has expected HTTP headers and valid key",
			Request: Request().
				Path("/ipns/{{name}}", ipnsV2).
				Query("format", "ipns-record"),
			Response: Expect().
				Headers(
					Header("Content-Disposition").Contains("attachment;"),
					Header("Content-Type").Contains("application/vnd.ipfs.ipns-record"),
					Header("Cache-Control").Contains("public, max-age=1800"),
				).
				Body(
					IsIPNSRecord(ipnsV2).
						IsValid().
						PointsTo("/ipfs/{{cid}}", cidIPNSV2),
				),
		},
		{
			Name: "GET IPNS Record (V1+V2) with 'Accept: application/vnd.ipfs.ipns-record' has expected HTTP headers and valid key",
			Request: Request().
				Path("/ipns/{{name}}", ipnsV1V2).
				Header("Accept", "application/vnd.ipfs.ipns-record"),
			Response: Expect().
				Headers(
					Header("Content-Disposition").Contains("attachment;"),
					Header("Content-Type").Contains("application/vnd.ipfs.ipns-record"),
					Header("Cache-Control").Contains("public, max-age=1800"),
				).
				Body(
					IsIPNSRecord(ipnsV1V2).
						IsValid().
						PointsTo("/ipfs/{{cid}}", cidIPNSV1V2),
				),
		},
		{
			Name: "GET IPNS Record (V2) with 'Accept: application/vnd.ipfs.ipns-record' has expected HTTP headers and valid key",
			Request: Request().
				Path("/ipns/{{name}}", ipnsV2).
				Header("Accept", "application/vnd.ipfs.ipns-record"),
			Response: Expect().
				Headers(
					Header("Content-Disposition").Contains("attachment;"),
					Header("Content-Type").Contains("application/vnd.ipfs.ipns-record"),
					Header("Cache-Control").Contains("public, max-age=1800"),
				).
				Body(
					IsIPNSRecord(ipnsV2).
						IsValid().
						PointsTo("/ipfs/{{cid}}", cidIPNSV2),
				),
		},
		{
			Name: "GET IPNS Record with explicit ?filename= succeeds with modified Content-Disposition header",
			Request: Request().
				Path("/ipns/{{name}}", ipnsV1V2).
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
