package main

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/ipns"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewayIPNSRecords(t *testing.T) {
	// Test HTTP Gateway IPNS Record (application/vnd.ipfs.ipns-record) Support

	// fixture := car.MustOpenUnixfsCar("t0124/fixtures.car") // TODO: implement the check for a CID
	ipnsRecord := ipns.MustOpenRecord("t0124/key1.ipns-record")

	tests := SugarTests{
		{
			Name: "GET KEY with format=ipns-record, validate key, and headers",
			Request: Request().
				Url("ipns/%s", ipnsRecord.Key()).
				Query("format", "ipns-record"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").
						Equals("application/vnd.ipfs.ipns-record"),
					Header("Content-Disposition").
						Contains("attachment;"),
					Header("Cache-Control").
						Matches("public, max-age=%d", ipnsRecord.TTL()),
				),
			Body(
				func(b []byte) bool {
					return ipnsRecord.Verify(b)
				},
			),
			// TODO: implement key verify + "should contain CID"
			// https://github.com/ipfs/kubo/blob/bc972a25851557ef2db193504325fcbe05a9508f/test/sharness/t0124-gateway-ipns-record.sh#L27
		},
		{
			Name: "GET KEY with 'Accept: application/vnd.ipfs.ipns-record', validate key, and headers",
			Request: Request().
				Url("ipns/%s", ipnsRecord.Key()).
				Headers(
					Header("Accept", "application/vnd.ipfs.ipns-record"),
				),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").
						Equals("application/vnd.ipfs.ipns-record"),
					Header("Content-Disposition").
						Contains("attachment;"),
					Header("Cache-Control").
						Matches("public, max-age=%d", ipnsRecord.TTL()),
				),
			Body(
				func(b []byte) bool {
					return ipnsRecord.Verify(b)
				},
			),
			// TODO: implement key verify + "should contain CID"
			// https://github.com/ipfs/kubo/blob/bc972a25851557ef2db193504325fcbe05a9508f/test/sharness/t0124-gateway-ipns-record.sh#L27
		},
		{
			Name: "GET KEY with expliciy ?filename= succeeds with modified Content-Disposition header",
			Request: Request().
				Url("ipns/%s", ipnsRecord.Key()).
				Query("filename", "testтест.ipns-record").
				Query("format", "ipns-record"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Disposition").
						Equals("attachment; filename=\"test____.ipns-record\"; filename*=UTF-8'''\"testтест.ipns-record\""),
					Header("Content-Type").
						Equals("application/vnd.ipfs.ipns-record"),
				),
		},
	}.Build()

	Run(t, tests)
}
