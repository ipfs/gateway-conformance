package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/ipns"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

// TODO(laurent): this was t0124_gateway_ipns_record_test.go

func TestGatewayIPNSRecord(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0124/fixtures.car")
	file := fixture.MustGetRoot()
	fileCID := file.Cid()

	// test_expect_success "Add the test directory & IPNS records" '
	// ipfs dag import ../t0124-gateway-ipns-record/fixtures.car &&
	// ipfs routing put /ipns/${IPNS_KEY} ../t0124-gateway-ipns-record/${IPNS_KEY}.ipns-record
	// '
	ipns := MustOpenIPNSRecordWithKey("t0124/k51qzi5uqu5dh71qgwangrt6r0nd4094i88nsady6qgd1dhjcyfsaqmpp143ab.ipns-record")
	ipnsName := ipns.Key()

	tests := SugarTests{
		// test_expect_success "Create and Publish IPNS Key" '
		// curl "http://127.0.0.1:$GWAY_PORT/ipns/$IPNS_KEY" > curl_output_filename &&
		// test_should_contain "Hello IPFS" curl_output_filename
		// '
		{
			Name: "GET an IPNS record from the gateway",
			Request: Request().
				Path("/ipns/{{name}}", ipnsName),
			Response: Expect().
				Body(file.RawData()),
		},
		// test_expect_success "GET KEY with format=ipns-record and validate key" '
		// curl "http://127.0.0.1:$GWAY_PORT/ipns/$IPNS_KEY?format=ipns-record" > curl_output_filename &&
		// ipfs name inspect --verify $IPNS_KEY < curl_output_filename > verify_output &&
		// test_should_contain "$FILE_CID" verify_output
		// '
		// test_expect_success "GET KEY with format=ipns-record has expected HTTP headers" '
		// curl -sD - "http://127.0.0.1:$GWAY_PORT/ipns/$IPNS_KEY?format=ipns-record" > curl_output_filename 2>&1 &&
		// test_should_contain "Content-Disposition: attachment;" curl_output_filename &&
		// test_should_contain "Content-Type: application/vnd.ipfs.ipns-record" curl_output_filename &&
		// test_should_contain "Cache-Control: public, max-age=3155760000" curl_output_filename
		// '
		{
			Name: "GET KEY with format=ipns-record has expected HTTP headers and valid key",
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
		// test_expect_success "GET KEY with 'Accept: application/vnd.ipfs.ipns-record' and validate key" '
		// curl -H "Accept: application/vnd.ipfs.ipns-record" "http://127.0.0.1:$GWAY_PORT/ipns/$IPNS_KEY" > curl_output_filename &&
		// ipfs name inspect --verify $IPNS_KEY < curl_output_filename > verify_output &&
		// test_should_contain "$FILE_CID" verify_output
		// '
		// test_expect_success "GET KEY with 'Accept: application/vnd.ipfs.ipns-record' has expected HTTP headers" '
		// curl -H "Accept: application/vnd.ipfs.ipns-record" -sD - "http://127.0.0.1:$GWAY_PORT/ipns/$IPNS_KEY" > curl_output_filename 2>&1 &&
		// test_should_contain "Content-Disposition: attachment;" curl_output_filename &&
		// test_should_contain "Content-Type: application/vnd.ipfs.ipns-record" curl_output_filename &&
		// test_should_contain "Cache-Control: public, max-age=3155760000" curl_output_filename
		// '
		{
			Name: "GET KEY with 'Accept: application/vnd.ipfs.ipns-record' has expected HTTP headers and valid key",
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
		// test_expect_success "GET KEY with expliciy ?filename= succeeds with modified Content-Disposition header" '
		// curl -sD - "http://127.0.0.1:$GWAY_PORT/ipns/$IPNS_KEY?format=ipns-record&filename=testтест.ipns-record" > curl_output_filename 2>&1 &&
		// grep -F "Content-Disposition: attachment; filename=\"test____.ipns-record\"; filename*=UTF-8'\'\''test%D1%82%D0%B5%D1%81%D1%82.ipns-record" curl_output_filename &&
		// test_should_contain "Content-Type: application/vnd.ipfs.ipns-record" curl_output_filename
		// '
		{
			Name: "GET KEY with expliciy ?filename= succeeds with modified Content-Disposition header",
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
