package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling"
	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/helpers"
	"github.com/ipfs/gateway-conformance/tooling/ipns"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
	. "github.com/ipfs/gateway-conformance/tooling/tmpl"
)

func TestGatewayJsonCbor(t *testing.T) {
	tooling.LogTestGroup(t, GroupJSONCbor)

	fixture := car.MustOpenUnixfsCar("path_gateway_dag/gateway-json-cbor.car")

	fileJSON := fixture.MustGetNode("ą", "ę", "t.json")
	fileJSONCID := fileJSON.Cid()
	fileJSONData := fileJSON.RawData()

	tests := SugarTests{
		{
			Name: "GET UnixFS file with JSON bytes is returned with application/json Content-Type - without headers",
			Hint: `
			## Quick regression check for JSON stored on UnixFS:
			## it has nothing to do with DAG-JSON and JSON codecs,
			## but a lot of JSON data is stored on UnixFS and is requested with or without various hints
			## and we want to avoid surprises like https://github.com/protocol/bifrost-infra/issues/2290
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", fileJSONCID),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").
						Equals("application/json"),
				).
				Body(fileJSONData),
		},
		{
			Name: "GET UnixFS file with JSON bytes is returned with application/json Content-Type - with headers",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#accept-request-header",
			Hint: `
			## Quick regression check for JSON stored on UnixFS:
			## it has nothing to do with DAG-JSON and JSON codecs,
			## but a lot of JSON data is stored on UnixFS and is requested with or without various hints
			## and we want to avoid surprises like https://github.com/protocol/bifrost-infra/issues/2290
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", fileJSONCID).
				Headers(
					Header("Accept", "application/json"),
				),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").
						Equals("application/json"),
				).
				Body(fileJSONData),
		},
		{
			Name: "GET raw block with JSON bytes prefers format over Accept header",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#format-request-query-parameter",
			Hint: `
			Per IPIP-0523, the format query parameter should be preferred over the
			Accept header when both are present. This test verifies that format=json
			overrides Accept: application/vnd.ipld.raw.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}?format=json", fileJSONCID).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
				),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").
						Equals("application/json"),
				).
				Body(fileJSONData),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayDAG)
}

// ## IPIP-0524: Codec mismatch returns 406 Not Acceptable
// ## Conversions between codecs are not part of the generic gateway specs.
// ## Requesting a format that doesn't match the block's codec returns 406.
// ##
// ## Implementations that support optional codec conversions for backward
// ## compatibility are free to skip these tests.
func TestCodecMismatchReturns406(t *testing.T) {
	tooling.LogTestGroup(t, GroupJSONCbor)

	// UnixFS (dag-pb) fixture
	unixfsFixture := car.MustOpenUnixfsCar("path_gateway_dag/gateway-json-cbor.car")
	unixfsFile := unixfsFixture.MustGetNode("ą", "ę", "file-źł.txt")
	unixfsFileCID := unixfsFile.Cid()

	// Native DAG-JSON fixture
	dagJSONFixture := car.MustOpenUnixfsCar("path_gateway_dag/dag-json-traversal.car").MustGetRoot()
	dagJSONCID := dagJSONFixture.Cid()

	// Native DAG-CBOR fixture
	dagCBORFixture := car.MustOpenUnixfsCar("path_gateway_dag/dag-cbor-traversal.car").MustGetRoot()
	dagCBORCID := dagCBORFixture.Cid()

	tests := SugarTests{
		// UnixFS (dag-pb) cannot be returned as dag-json or dag-cbor
		{
			Name: "GET UnixFS (dag-pb) file with format=dag-json returns 406 Not Acceptable",
			Hint: `
			IPIP-0524 clarifies that codec conversions are not part of the specs.
			Requesting dag-json for a dag-pb block returns 406.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", unixfsFileCID).
				Query("format", "dag-json"),
			Response: Expect().
				Status(406),
		},
		{
			Name: "GET UnixFS (dag-pb) file with format=dag-cbor returns 406 Not Acceptable",
			Hint: `
			IPIP-0524 clarifies that codec conversions are not part of the specs.
			Requesting dag-cbor for a dag-pb block returns 406.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", unixfsFileCID).
				Query("format", "dag-cbor"),
			Response: Expect().
				Status(406),
		},
		{
			Name: "GET UnixFS (dag-pb) file with Accept: application/vnd.ipld.dag-json returns 406 Not Acceptable",
			Hint: `
			IPIP-0524 clarifies that codec conversions are not part of the specs.
			Requesting dag-json for a dag-pb block returns 406.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", unixfsFileCID).
				Headers(
					Header("Accept", "application/vnd.ipld.dag-json"),
				),
			Response: Expect().
				Status(406),
		},
		// DAG-JSON cannot be returned as dag-cbor
		{
			Name: "GET DAG-JSON block with format=dag-cbor returns 406 Not Acceptable",
			Hint: `
			IPIP-0524 clarifies that codec conversions are not part of the specs.
			Requesting dag-cbor for a dag-json block returns 406.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", dagJSONCID).
				Query("format", "dag-cbor"),
			Response: Expect().
				Status(406),
		},
		// DAG-CBOR cannot be returned as dag-json
		{
			Name: "GET DAG-CBOR block with format=dag-json returns 406 Not Acceptable",
			Hint: `
			IPIP-0524 clarifies that codec conversions are not part of the specs.
			Requesting dag-json for a dag-cbor block returns 406.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", dagCBORCID).
				Query("format", "dag-json"),
			Response: Expect().
				Status(406),
		},
		{
			Name: "GET DAG-CBOR block with Accept: application/vnd.ipld.dag-json returns 406 Not Acceptable",
			Hint: `
			IPIP-0524 clarifies that codec conversions are not part of the specs.
			Requesting dag-json for a dag-cbor block returns 406.
			`,
			Request: Request().
				Path("/ipfs/{{cid}}", dagCBORCID).
				Headers(
					Header("Accept", "application/vnd.ipld.dag-json"),
				),
			Response: Expect().
				Status(406),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayDAG)
}

// # Requesting CID with plain json (0x0200) and cbor (0x51) codecs
// # (note these are not UnixFS, not DAG-* variants, just raw block identified by a CID with a special codec)
func TestPlainCodec(t *testing.T) {
	tooling.LogTestGroup(t, GroupJSONCbor)

	table := []struct {
		Name        string
		Format      string
		Disposition string
		Checker     func(value []byte) Check[[]byte]
	}{
		{"plain JSON codec", "json", "inline", IsJSONEqual},
		{"plain CBOR codec", "cbor", "attachment", IsEqualBytes},
	}

	for _, row := range table {
		plain := car.MustOpenUnixfsCar(Fmt("path_gateway_dag/plain-{{format}}.car", row.Format)).MustGetRoot()
		plainCID := plain.Cid()

		tests := SugarTests{}.
			Append(
				helpers.IncludeRandomRangeTests(t,
					SugarTest{
						Name: Fmt(`GET {{name}} without Accept or format= has expected "{{format}}" Content-Type and body as-is`, row.Name, row.Format),
						Hint: `
				No explicit format, just codec in CID
				`,
						Request: Request().
							Path("/ipfs/{{cid}}", plainCID),
						Response: Expect().
							Headers(
								Header("Content-Disposition").
									Contains(Fmt(`{{disposition}}; filename="{{cid}}.{{format}}"`, row.Disposition, plainCID, row.Format)),
							),
					},
					plain.RawData(),
					Fmt("application/{{format}}", row.Format),
				)...).
			Append(
				helpers.IncludeRandomRangeTests(t,
					SugarTest{
						Name: Fmt("GET {{name}} with ?format= has expected {{format}} Content-Type and body as-is", row.Name, row.Format),
						Hint: `
				Explicit format still gives correct output, just codec in CID
				`,
						Request: Request().
							Path("/ipfs/{{cid}}", plainCID).
							Query("format", row.Format),
						Response: Expect().
							Headers(
								Header("Content-Disposition").
									Contains(`{{disposition}}; filename="{{cid}}.{{format}}"`, row.Disposition, plainCID, row.Format),
							),
					},
					plain.RawData(),
					Fmt("application/{{format}}", row.Format),
				)...).
			Append(
				helpers.IncludeRandomRangeTests(t,
					SugarTest{
						Name: Fmt("GET {{name}} with Accept has expected {{format}} Content-Type and body as-is, with single range request", row.Name, row.Format),
						Hint: `
				Explicit format still gives correct output, just codec in CID
				`,
						Request: Request().
							Path("/ipfs/{{cid}}", plainCID).
							Headers(
								Header("Accept", Fmt("application/{{format}}", row.Format)),
							),
						Response: Expect().
							Headers(
								Header("Content-Disposition").
									Contains(`{{disposition}}; filename="{{cid}}.{{format}}"`, row.Disposition, plainCID, row.Format),
							),
					},
					plain.RawData(),
					Fmt("application/{{format}}", row.Format),
				)...)

		RunWithSpecs(t, tests, specs.PathGatewayDAG)
	}
}

// ## Pathing, traversal over DAG-JSON and DAG-CBOR
func TestPathing(t *testing.T) {
	tooling.LogTestGroup(t, GroupJSONCbor)

	dagJSONTraversal := car.MustOpenUnixfsCar("path_gateway_dag/dag-json-traversal.car").MustGetRoot()
	dagCBORTraversal := car.MustOpenUnixfsCar("path_gateway_dag/dag-cbor-traversal.car").MustGetRoot()

	dagJSONTraversalCID := dagJSONTraversal.Cid()
	dagCBORTraversalCID := dagCBORTraversal.Cid()

	tests := SugarTests{
		{
			Name: "GET DAG-JSON traversal returns 501 if there is path remainder",
			Request: Request().
				Path("/ipfs/{{cid}}/foo", dagJSONTraversalCID).
				Query("format", "dag-json"),
			Response: Expect().
				Status(501), // reading IPLD Kinds other than Links (CBOR Tag 42) is not implemented
		},
		{
			Name: "GET DAG-JSON traverses multiple links",
			Request: Request().
				Path("/ipfs/{{cid}}/foo/link/bar", dagJSONTraversalCID).
				Query("format", "dag-json"),
			Response: Expect().
				Status(200).
				Body(
					// TODO: I like that this text is readable and easy to understand.
					// 		 but we might prefer matching abstract values, something like "IsJSONEqual(someFixture.formatedAsJSON))"
					IsJSONEqual([]byte(`{"hello": "this is not a link"}`)),
				),
		},
		{
			Name: "GET DAG-JSON returns 404 on non-existing link",
			Request: Request().
				Path("/ipfs/{{cid}}/foo/i-do-not-exist", dagJSONTraversalCID),
			Response: Expect().
				Status(404),
		},
		{
			Name: "GET DAG-CBOR traversal returns 501 if there is path remainder",
			Request: Request().
				Path("/ipfs/{{cid}}/foo", dagCBORTraversalCID).
				Query("format", "dag-cbor"),
			Response: Expect().
				Status(501), // reading IPLD Kinds other than Links (CBOR Tag 42) is not implemented
		},
		{
			Name: "GET DAG-CBOR traverses multiple links",
			Request: Request().
				Path("/ipfs/{{cid}}/foo/link/bar", dagCBORTraversalCID).
				Query("format", "dag-cbor"),
			Response: Expect().
				Status(200).
				Body(
					// CBOR bytes for {"hello": "this is not a link"}
					IsEqualBytes([]byte{0xa1, 0x65, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x72, 0x74, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20, 0x6e, 0x6f, 0x74, 0x20, 0x61, 0x20, 0x6c, 0x69, 0x6e, 0x6b}),
				),
		},
		{
			Name: "GET DAG-CBOR returns 404 on non-existing link",
			Request: Request().
				Path("/ipfs/{{cid}}/foo/i-do-not-exist", dagCBORTraversalCID),
			Response: Expect().
				Status(404),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayDAG)
}

// ## NATIVE TESTS for DAG-JSON (0x0129) and DAG-CBOR (0x71):
// ## DAG- regression tests for core behaviors when native DAG-(CBOR|JSON) is requested
func TestNativeDag(t *testing.T) {
	tooling.LogTestGroup(t, GroupJSONCbor)

	missingCID := car.RandomCID()

	table := []struct {
		Name        string
		Format      string
		Disposition string
		Checker     func(value []byte) Check[[]byte]
	}{
		{"plain JSON codec", "json", "inline", IsJSONEqual},
		{"plain CBOR codec", "cbor", "attachment", IsEqualBytes},
	}

	for _, row := range table {
		dagTraversal := car.MustOpenUnixfsCar(Fmt("path_gateway_dag/dag-{{format}}-traversal.car", row.Format)).MustGetRoot()
		dagTraversalCID := dagTraversal.Cid()
		formatted := dagTraversal.Formatted("dag-" + row.Format)

		tests := SugarTests{
			{
				Name: Fmt("GET {{name}} from /ipfs without explicit format returns the same payload as the raw block", row.Name),
				Hint: `GET without explicit format and Accept: text/html returns raw block`,
				Request: Request().
					Path("/ipfs/{{cid}}", dagTraversalCID),
				Response: Expect().
					Status(200).
					Body(
						row.Checker(formatted),
					),
			},
			{
				Name: Fmt("GET {{name}} from /ipfs with format=dag-{{format}} returns the same payload as the raw block", row.Name, row.Format),
				Hint: `GET dag-cbor block via Accept and ?format and ensure both are the same as ipfs block get output`,
				Request: Request().
					Path("/ipfs/{{cid}}", dagTraversalCID).
					Query("format", Fmt("dag-{{format}}", row.Format)),
				Response: Expect().
					Status(200).
					Body(
						row.Checker(formatted),
					),
			},
			{
				Name: Fmt("GET {{name}} from /ipfs with application/vnd.ipld.dag-{{format}} returns the same payload as the raw block", row.Name, row.Format),
				Request: Request().
					Path("/ipfs/{{cid}}", dagTraversalCID).
					Header("Accept", Fmt("application/vnd.ipld.dag-{{format}}", row.Format)),
				Response: Expect().
					Status(200).
					Body(
						row.Checker(formatted),
					),
			},
			{
				Name: Fmt("GET {{name}} with format={{format}} returns same payload as format=dag-{{format}} but with plain Content-Type", row.Name, row.Format),
				Hint: `Make sure DAG-* can be requested as plain JSON or CBOR and response has plain Content-Type for interop purposes`,
				Request: Request().
					Path("/ipfs/{{cid}}", dagTraversalCID).
					Query("format", row.Format),
				Response: Expect().
					Status(200).
					Header(Header("Content-Type", "application/{{format}}", row.Format)).
					Body(
						row.Checker(formatted),
					),
			},
			{
				Name: Fmt("GET {{name}} with Accept: application/{{format}} returns same payload as application/vnd.ipld.dag-{{format}} but with plain Content-Type", row.Name, row.Format),
				Request: Request().
					Path("/ipfs/{{cid}}", dagTraversalCID).
					Header("Accept", "application/{{format}}", row.Format),
				Response: Expect().
					Status(200).
					Header(Header("Content-Type", "application/{{format}}", row.Format)).
					Body(
						row.Checker(formatted),
					),
			},
			{
				Name: Fmt("GET response for application/vnd.ipld.dag-{{format}} has expected Content-Type", row.Format),
				Hint: `Make sure expected HTTP headers are returned with the dag- block`,
				Request: Request().
					Path("/ipfs/{{cid}}", dagTraversalCID).
					Header("Accept", Fmt("application/vnd.ipld.dag-{{format}}", row.Format)),
				Response: Expect().
					Headers(
						Header("Content-Type").Hint("expected Content-Type").Equals("application/vnd.ipld.dag-{{format}}", row.Format),
						Header("Content-Length").Spec("https://specs.ipfs.tech/http-gateways/path-gateway/#content-disposition-response-header").Hint("includes Content-Length").Equals("{{length}}", len(dagTraversal.RawData())),
						Header("Content-Disposition").Hint("includes Content-Disposition").Contains(`{{disposition}}; filename="{{cid}}.{{format}}"`, row.Disposition, dagTraversalCID, row.Format),
						Header("X-Content-Type-Options").Hint("includes nosniff hint").Contains("nosniff"),
					),
			},
			{
				Name: Fmt("GET for application/vnd.ipld.dag-{{format}} with query filename includes Content-Disposition with custom filename", row.Format),
				Request: Request().
					Path("/ipfs/{{cid}}", dagTraversalCID).
					Query("filename", Fmt("foobar.{{format}}", row.Format)).
					Header("Accept", Fmt("application/vnd.ipld.dag-{{format}}", row.Format)),
				Response: Expect().
					Headers(
						Header("Content-Disposition").
							Hint("includes Content-Disposition").
							Contains(`{{disposition}}; filename="foobar.{{format}}"`, row.Disposition, row.Format),
					),
			},
			{
				Name: Fmt("GET for application/vnd.ipld.dag-{{format}} with ?download=true forces Content-Disposition: attachment", row.Format),
				Request: Request().
					Path("/ipfs/{{cid}}", dagTraversalCID).
					Query("filename", Fmt("foobar.{{format}}", row.Format)).
					Query("download", "true").
					Header("Accept", Fmt("application/vnd.ipld.dag-{{format}}", row.Format)),
				Response: Expect().
					Headers(
						Header("Content-Disposition").
							Hint("includes Content-Disposition").
							Contains(`attachment; filename="foobar.{{format}}"`, row.Format),
					),
			},
			{
				Name: Fmt("Cache control HTTP headers ({{format}})", row.Format),
				Hint: `(basic checks, detailed behavior is tested in t0116-gateway-cache.sh)`,
				Request: Request().
					Path("/ipfs/{{cid}}", dagTraversalCID).
					Header("Accept", Fmt("application/vnd.ipld.dag-{{format}}", row.Format)),
				Response: Expect().
					Headers(
						Header("Etag").Hint("includes Etag").Contains("{{cid}}.dag-{{format}}", dagTraversalCID, row.Format),
						Header("X-Ipfs-Path").Hint("includes X-Ipfs-Path").Exists(),
						Header("X-Ipfs-Roots").Hint("includes X-Ipfs-Roots").Exists(),
						Header("Cache-Control").Hint("includes Cache-Control").Contains("public, max-age=29030400, immutable"),
					),
			},
			{
				Name: Fmt("HEAD {{name}} with no explicit format returns HTTP 200", row.Name),
				Request: Request().
					Path("/ipfs/{{cid}}", dagTraversalCID).
					Method("HEAD"),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Type").Hint("includes Content-Type").Contains("application/vnd.ipld.dag-{{format}}", row.Format),
						Header("Content-Length").Hint("includes Content-Length").Exists(),
					),
			},
			{
				Name: Fmt("HEAD {{name}} with an explicit format returns HTTP 200", row.Name),
				Request: Request().
					Path("/ipfs/{{cid}}", dagTraversalCID).
					Query("format", Fmt("dag-{{format}}", row.Format)).
					Method("HEAD"),
				Response: Expect().
					Status(200).
					Headers(
						Header("Etag").Hint("includes Etag").Contains("{{cid}}.dag-{{format}}", dagTraversalCID, row.Format),
						Header("Content-Type").Hint("includes Content-Type").Contains("application/vnd.ipld.dag-{{format}}", row.Format),
						Header("Content-Length").Hint("includes Content-Length").Exists(),
					),
			},
			{
				Name: Fmt("HEAD {{name}} with only-if-cached for missing block returns HTTP 412 Precondition Failed", row.Name),
				Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#only-if-cached",
				Request: Request().
					Path("/ipfs/{{cid}}", missingCID).
					Header("Cache-Control", "only-if-cached").
					Method("HEAD"),
				Response: Expect().
					Status(412),
			},
			{
				Name: Fmt("GET {{name}} on /ipfs with Accept: text/html returns HTML (dag-index-html)", row.Name),
				Request: Request().
					Path("/ipfs/{{cid}}/", dagTraversalCID).
					Header("Accept", "text/html"),
				Response: AllOf(
					Expect().
						Status(200).
						Headers(
							Header("Etag").Contains("DagIndex-"),
							Header("Content-Type").Contains("text/html"),
							Header("Content-Disposition").IsEmpty(),
						).
						Body(
							Contains("</html>"),
						),
					AnyOf(
						Expect().Headers(Header("Cache-Control").IsEmpty()),
						Expect().Headers(Header("Cache-Control").Equals("public, max-age=604800, stale-while-revalidate=2678400")),
					),
				),
			},
		}
		tests.Append(helpers.OnlyRandomRangeTests(t,
			SugarTest{
				Name: Fmt("GET {{name}} on /ipfs with no explicit header", row.Name),
				Request: Request().
					Path("/ipfs/{{cid}}/", dagTraversalCID),
				Response: Expect(),
			},
			dagTraversal.RawData(), Fmt("application/vnd.ipld.dag-{{format}}", row.Format),
		)...).Append(
			helpers.OnlyRandomRangeTests(t,
				SugarTest{
					Name: Fmt("GET {{name}} on /ipfs with dag content headers", row.Name),
					Request: Request().
						Path("/ipfs/{{cid}}/", dagTraversalCID).
						Headers(
							Header("Accept", "application/vnd.ipld.dag-{{format}}", row.Format),
						),
					Response: Expect(),
				},
				dagTraversal.RawData(),
				Fmt("application/vnd.ipld.dag-{{format}}", row.Format),
			)...).Append(
			helpers.OnlyRandomRangeTests(t,
				SugarTest{
					Name: Fmt("GET {{name}} on /ipfs with non-dag content headers", row.Name),
					Request: Request().
						Path("/ipfs/{{cid}}/", dagTraversalCID).
						Headers(
							Header("Accept", "application/{{format}}", row.Format),
						),
					Response: Expect(),
				},
				dagTraversal.RawData(),
				Fmt("application/{{format}}", row.Format),
			)...)

		RunWithSpecs(t, tests, specs.PathGatewayDAG)
	}

	// Test that DAG-CBOR can be rendered as HTML. This is not codec conversion,
	// but a human-readable preview for browsing. Unlike codec conversions which
	// were removed by IPIP-0524, HTML rendering remains part of the gateway spec.
	dagCborFixture := car.MustOpenUnixfsCar("path_gateway_dag/dag-cbor-traversal.car").MustGetRoot()
	dagCborCID := dagCborFixture.Cid()
	RunWithSpecs(t, SugarTests{
		SugarTest{
			Name: "GET DAG-CBOR with Accept: text/html returns HTML preview",
			Hint: "text/html returns a human-readable representation of the data",
			Request: Request().
				Path("/ipfs/{{cid}}/", dagCborCID).
				Headers(
					Header("Accept", "text/html"),
				),
			Response: Expect().Body(Contains("</html>")),
		},
	}, specs.PathGatewayDAG)
}

func TestGatewayJSONCborAndIPNS(t *testing.T) {
	tooling.LogTestGroup(t, GroupIPNS)

	ipnsIdDagJSON := "k51qzi5uqu5dhjghbwdvbo6mi40htrq6e2z4pwgp15pgv3ho1azvidttzh8yy2"
	ipnsIdDagCBOR := "k51qzi5uqu5dghjous0agrwavl8vzl64xckoqzwqeqwudfr74kfd11zcyk3b7l"

	ipnsDagJSON := ipns.MustOpenIPNSRecordWithKey(Fmt("path_gateway_dag/{{id}}.ipns-record", ipnsIdDagJSON))
	ipnsDagCBOR := ipns.MustOpenIPNSRecordWithKey(Fmt("path_gateway_dag/{{id}}.ipns-record", ipnsIdDagCBOR))

	table := []struct {
		Name    string
		Format  string
		fixture *ipns.IpnsRecord
	}{
		{"plain JSON codec", "json", ipnsDagJSON},
		{"plain CBOR codec", "cbor", ipnsDagCBOR},
	}

	tests := SugarTests{}

	for _, row := range table {
		plain := car.MustOpenUnixfsCar(Fmt("path_gateway_dag/dag-{{format}}-traversal.car", row.Format)).MustGetRoot()
		plainCID := plain.Cid()

		// # IPNS behavior (should be same as immutable /ipfs, but with different caching headers)
		// # To keep tests small we only confirm payload is the same, and then only test delta around caching headers.
		tests = append(tests, SugarTests{
			{
				Name: Fmt("GET {{name}} from /ipns without explicit format returns the same payload as /ipfs", row.Name),
				Requests: Requests(
					Request().
						Path("/ipfs/{{cid}}", plainCID),
					Request().
						Path("/ipns/{{id}}", row.fixture.Key()),
				),
				Responses: Responses().
					HaveTheSamePayload(),
			},
			{
				Name: Fmt("GET {{name}} from /ipns with explicit format returns the same payload as /ipfs", row.Name),
				Requests: Requests(
					Request().
						Path("/ipfs/{{cid}}", plainCID).
						Query("format", "dag-{{format}}", row.Format),
					Request().
						Path("/ipns/{{id}}", row.fixture.Key()).
						Query("format", "dag-{{format}}", row.Format),
				),
				Responses: Responses().
					HaveTheSamePayload(),
			},
			{
				Name: Fmt("GET {{name}} from /ipns with explicit application/vnd.ipld.dag-{{format}} has expected headers", row.Name, row.Format),
				Request: Request().
					Path("/ipns/{{id}}", row.fixture.Key()).
					Header("Accept", "application/vnd.ipld.dag-{{format}}", row.Format),
				Response: Expect().
					Headers(
						Header("Content-Type").Equals("application/vnd.ipld.dag-{{format}}", row.Format),
						Header("Etag").Equals(`"{{cid}}.dag-{{format}}"`, plainCID, row.Format),
						Header("X-Ipfs-Path").Not().IsEmpty(),
						Header("X-Ipfs-Roots").Not().IsEmpty(),
					),
			},
			{
				Name: Fmt("GET {{name}} on /ipns with Accept: text/html returns HTML (dag-index-html)", row.Name),
				Request: Request().
					Path("/ipns/{{id}}/", row.fixture.Key()).
					Header("Accept", "text/html"),
				Response: AllOf(
					Expect().
						Headers(
							Header("Etag").Contains("DagIndex-"),
							Header("Content-Type").Contains("text/html"),
							Header("Content-Disposition").IsEmpty(),
						).Body(
						Contains("</html>"),
					),
					AnyOf(
						Expect().Headers(Header("Cache-Control").IsEmpty()),
						Expect().Headers(Header("Cache-Control").Matches("public, max-age=*")),
					),
				),
			},
		}...)
	}

	RunWithSpecs(t, tests, specs.PathGatewayDAG, specs.PathGatewayIPNS)
}
