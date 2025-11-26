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
			Name: "GET raw block with JSON bytes prefers format over accept header",
			Spec: "https://specs.ipfs.tech/http-gateways/path-gateway/#format-request-query-parameter",
			Hint: `
			## The format search query parameter should be preferred over the Accept
			## header when both are present
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

// ## Reading UnixFS (data encoded with dag-pb codec) as DAG-CBOR and DAG-JSON
// ## (returns representation defined in https://ipld.io/specs/codecs/dag-pb/spec/#logical-format)
func TestDagPbConversion(t *testing.T) {
	tooling.LogTestGroup(t, GroupJSONCbor)

	fixture := car.MustOpenUnixfsCar("path_gateway_dag/gateway-json-cbor.car")

	dir := fixture.MustGetRoot()
	file := fixture.MustGetNode("ą", "ę", "file-źł.txt")

	dirCID := dir.Cid()
	fileCID := file.Cid()
	fileData := file.RawData()

	table := []struct {
		Name        string
		Format      string
		Disposition string
	}{
		{"DAG-JSON", "json", "inline"},
		{"DAG-CBOR", "cbor", "attachment"},
	}

	for _, row := range table {
		// ipfs dag get --output-codec dag-$format $FILE_CID > ipfs_dag_get_output
		formatedFile := file.Formatted("dag-" + row.Format)
		formatedDir := dir.Formatted("dag-" + row.Format)

		tests := SugarTests{
			{
				Name: Fmt("GET UnixFS file as {{name}} with format=dag-{{format}} converts to the expected Content-Type", row.Name, row.Format),
				Request: Request().
					Path("/ipfs/{{cid}}", fileCID).
					Query("format", "dag-"+row.Format),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Type").
							Equals("application/vnd.ipld.dag-{{format}}", row.Format),
						Header("Content-Disposition").
							Contains(`{{disposition}}; filename="{{cid}}.{{format}}"`, row.Disposition, fileCID, row.Format),
						Header("Content-Type").
							Not().Contains("application/{{format}}", row.Format),
					).Body(
					formatedFile,
				),
			},
			{
				Name: Fmt("GET UnixFS directory as {{name}} with format=dag-{{format}} converts to the expected Content-Type", row.Name, row.Format),
				Request: Request().
					Path("/ipfs/{{cid}}?format=dag-{{format}}", dirCID, row.Format),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Type").
							Equals("application/vnd.ipld.dag-{{format}}", row.Format),
						Header("Content-Disposition").
							Contains(`{{disposition}}; filename="{{cid}}.{{format}}"`, row.Disposition, dirCID, row.Format),
						Header("Content-Type").
							Not().Contains("application/{{format}}", row.Format),
					).Body(
					formatedDir,
				),
			},
			{
				Name: Fmt("GET UnixFS as {{name}} with 'Accept: application/vnd.ipld.dag-{{format}}' converts to the expected Content-Type", row.Name, row.Format),
				Request: Request().
					Path("/ipfs/{{cid}}", fileCID).
					Headers(
						Header("Accept", "application/vnd.ipld.dag-{{format}}", row.Format),
					),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Disposition").
							Contains(`{{disposition}}; filename="{{cid}}.{{format}}"`, row.Disposition, fileCID, row.Format),
						Header("Content-Type").
							Equals("application/vnd.ipld.dag-{{format}}", row.Format),
						Header("Content-Type").
							Not().Contains("application/{{format}}", row.Format),
					),
			},
			{
				Name: Fmt("GET UnixFS as {{name}} with 'Accept: foo, application/vnd.ipld.dag-{{format}},bar' converts to the expected Content-Type", row.Name, row.Format),
				Request: Request().
					Path("/ipfs/{{cid}}", fileCID).
					Headers(
						Header("Accept", "foo, application/vnd.ipld.dag-{{format}},bar", row.Format),
					),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Type").
							Equals("application/vnd.ipld.dag-{{format}}", row.Format),
					),
			},
			{
				Name: Fmt("GET UnixFS with format={{format}} (not dag-{{format}}) is no-op (no conversion)", row.Format),
				Request: Request().
					Path("/ipfs/{{cid}}?format={{format}}", fileCID, row.Format),
				Response: Expect().
					Status(200).
					Headers(
						// NOTE: kubo gateway returns "text/plain; charset=utf-8" for example
						Header("Content-Type").
							Contains("text/plain"),
						Header("Content-Type").
							Not().Contains("application/{{format}}", row.Format),
						Header("Content-Type").
							Not().Contains("application/vnd.ipld.dag-{{format}}", row.Format),
					).Body(
					fileData,
				),
			},
			{
				Name: Fmt("GET UnixFS with 'Accept: application/{{format}}' (not dag-{{format}}) is no-op (no conversion)", row.Format),
				Request: Request().
					Path("/ipfs/{{cid}}", fileCID).
					Headers(
						Header("Accept", "application/{{format}}", row.Format),
					),
				Response: Expect().
					Status(200).
					Headers(
						// NOTE: kubo gateway returns "text/plain; charset=utf-8" for example
						Header("Content-Type").
							Contains("text/plain"),
						Header("Content-Type").
							Not().Contains("application/{{format}}", row.Format),
						Header("Content-Type").
							Not().Contains("application/vnd.ipld.dag-{{format}}", row.Format),
					).Body(
					fileData,
				),
			},
		}

		RunWithSpecs(t, tests, specs.PathGatewayDAG)
	}
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
		plainOrDag := car.MustOpenUnixfsCar(Fmt("path_gateway_dag/plain-cbor-that-can-be-dag-{{format}}.car", row.Format)).MustGetRoot()
		formatted := plainOrDag.Formatted("dag-" + row.Format)

		plainCID := plain.Cid()
		plainOrDagCID := plainOrDag.Cid()

		var dagFormattedResponse []byte

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
				)...).
			Append(
				SugarTest{
					Name: Fmt("GET {{name}} with format=dag-{{format}} interprets {{format}} as dag-* variant and produces expected Content-Type and body", row.Name, row.Format),
					Hint: `
				Explicit dag-* format passed, attempt to parse as dag* variant
				Note: this works only for simple JSON that can be upgraded to  DAG-JSON.
				`,
					Request: Request().
						Path("/ipfs/{{cid}}", plainOrDagCID).
						Query("format", Fmt("dag-{{format}}", row.Format)),
					Response: Expect().
						Status(200).
						Headers(
							Header("Content-Disposition").
								Contains(`{{disposition}}; filename="{{cid}}.{{format}}"`, row.Disposition, plainOrDagCID, row.Format),
							Header("Content-Type").
								Contains("application/vnd.ipld.dag-{{format}}", row.Format),
						).Body(
						Checks("", func(t []byte) bool {
							innerCheck := row.Checker(formatted).Check(t)
							if innerCheck.Success {
								dagFormattedResponse = t
								return true
							}
							return false
						}),
					),
				},
			)

		RunWithSpecs(t, tests, specs.PathGatewayDAG)

		if dagFormattedResponse != nil {
			rangeTests := helpers.OnlyRandomRangeTests(t,
				SugarTest{
					Name: Fmt("GET {{name}} with format=dag-{{format}} interprets {{format}} as dag-* variant and produces expected Content-Type and body, with single range request", row.Name, row.Format),
					Hint: `
				Explicit dag-* format passed, attempt to parse as dag* variant
				Note: this works only for simple JSON that can be upgraded to  DAG-JSON.
				`,
					Request: Request().
						Path("/ipfs/{{cid}}", plainOrDagCID).
						Query("format", Fmt("dag-{{format}}", row.Format)),
					Response: Expect().
						Headers(
							Header("Content-Disposition").
								Contains(`{{disposition}}; filename="{{cid}}.{{format}}"`, row.Disposition, plainOrDagCID, row.Format),
						),
				},
				dagFormattedResponse,
				Fmt("application/vnd.ipld.dag-{{format}}", row.Format),
			)
			RunWithSpecs(t, rangeTests, specs.PathGatewayDAG)
		}
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
				Name: Fmt("HEAD {{name}} with an explicit DAG-JSON format returns HTTP 200", row.Name),
				Request: Request().
					Path("/ipfs/{{cid}}", dagTraversalCID).
					Query("format", "dag-json").
					Method("HEAD"),
				Response: Expect().
					Status(200).
					Headers(
						Header("Etag").Hint("includes Etag").Contains("{{cid}}.dag-json", dagTraversalCID),
						Header("Content-Type").Hint("includes Content-Type").Contains("application/vnd.ipld.dag-json"),
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

	dagCborFixture := car.MustOpenUnixfsCar("path_gateway_dag/dag-cbor-traversal.car").MustGetRoot()
	dagCborCID := dagCborFixture.Cid()
	var dagJsonConvertedData []byte
	RunWithSpecs(t, SugarTests{
		SugarTest{
			Name: "Convert application/vnd.ipld.dag-cbor to application/vnd.ipld.dag-json",
			Hint: "",
			Request: Request().
				Path("/ipfs/{{cid}}/", dagCborCID).
				Headers(
					Header("Accept", "application/vnd.ipld.dag-json"),
				),
			Response: Expect().Body(Checks("", func(t []byte) bool {
				innerCheck := IsJSONEqual(dagCborFixture.Formatted("dag-json")).Check(t)
				if innerCheck.Success {
					dagJsonConvertedData = t
					return true
				}
				return false
			})),
		},
	}, specs.PathGatewayDAG)

	if dagJsonConvertedData != nil {
		rangeTests := helpers.OnlyRandomRangeTests(
			t,
			SugarTest{
				Name: "Convert application/vnd.ipld.dag-cbor to application/vnd.ipld.dag-json with range request includes correct bytes",
				Hint: "",
				Request: Request().
					Path("/ipfs/{{cid}}/", dagCborCID).
					Headers(
						Header("Accept", "application/vnd.ipld.dag-json"),
					),
				Response: Expect(),
			},
			dagJsonConvertedData,
			"application/vnd.ipld.dag-json")

		RunWithSpecs(t, rangeTests, specs.PathGatewayDAG)
	}

	RunWithSpecs(t, SugarTests{
		SugarTest{
			Name: "Convert application/vnd.ipld.dag-cbor to text/html",
			Hint: "",
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
