package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/ipns"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
	. "github.com/ipfs/gateway-conformance/tooling/tmpl"
)

// TODO(laurent): this was t0123_gateway_json_cbor_test

func TestGatewayJsonCbor(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0123-gateway-json-cbor.car")

	fileJSON := fixture.MustGetNode("ą", "ę", "t.json")
	fileJSONCID := fileJSON.Cid()
	fileJSONData := fileJSON.RawData()

	tests := SugarTests{
		{
			Name: "GET UnixFS file with JSON bytes is returned with application/json Content-Type (1)",
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
			Name: "GET UnixFS file with JSON bytes is returned with application/json Content-Type (2)",
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
	}

	RunWithSpecs(t, tests, specs.PathGatewayDAG)
}

// ## Reading UnixFS (data encoded with dag-pb codec) as DAG-CBOR and DAG-JSON
// ## (returns representation defined in https://ipld.io/specs/codecs/dag-pb/spec/#logical-format)
func TestDAgPbConversion(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0123-gateway-json-cbor.car")

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
			/**
				test_expect_success "GET UnixFS file as $name with format=dag-$format converts to the expected Content-Type" '
				curl -sD headers "http://127.0.0.1:$GWAY_PORT/ipfs/$FILE_CID?format=dag-$format" > curl_output 2>&1 &&
				ipfs dag get --output-codec dag-$format $FILE_CID > ipfs_dag_get_output 2>&1 &&
				test_cmp ipfs_dag_get_output curl_output &&
				test_should_contain "Content-Type: application/vnd.ipld.dag-$format" headers &&
				test_should_contain "Content-Disposition: ${disposition}\; filename=\"${FILE_CID}.${format}\"" headers &&
				test_should_not_contain "Content-Type: application/$format" headers
			'
			*/
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
			/**
			test_expect_success "GET UnixFS directory as $name with format=dag-$format converts to the expected Content-Type" '
			curl -sD headers "http://127.0.0.1:$GWAY_PORT/ipfs/$DIR_CID?format=dag-$format" > curl_output 2>&1 &&
			ipfs dag get --output-codec dag-$format $DIR_CID > ipfs_dag_get_output 2>&1 &&
			test_cmp ipfs_dag_get_output curl_output &&
			test_should_contain "Content-Type: application/vnd.ipld.dag-$format" headers &&
			test_should_contain "Content-Disposition: ${disposition}\; filename=\"${DIR_CID}.${format}\"" headers &&
			test_should_not_contain "Content-Type: application/$format" headers
			'
			*/
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
			/**
			test_expect_success "GET UnixFS as $name with 'Accept: application/vnd.ipld.dag-$format' converts to the expected Content-Type" '
			curl -sD - -H "Accept: application/vnd.ipld.dag-$format" "http://127.0.0.1:$GWAY_PORT/ipfs/$FILE_CID" > curl_output 2>&1 &&
			test_should_contain "Content-Disposition: ${disposition}\; filename=\"${FILE_CID}.${format}\"" curl_output &&
			test_should_contain "Content-Type: application/vnd.ipld.dag-$format" curl_output &&
			test_should_not_contain "Content-Type: application/$format" curl_output
			'
			*/
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
			/**
			test_expect_success "GET UnixFS as $name with 'Accept: foo, application/vnd.ipld.dag-$format,bar' converts to the expected Content-Type" '
			curl -sD - -H "Accept: foo, application/vnd.ipld.dag-$format,text/plain" "http://127.0.0.1:$GWAY_PORT/ipfs/$FILE_CID" > curl_output 2>&1 &&
			test_should_contain "Content-Type: application/vnd.ipld.dag-$format" curl_output
			'
			*/
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
			/**
			test_expect_success "GET UnixFS with format=$format (not dag-$format) is no-op (no conversion)" '
			curl -sD headers "http://127.0.0.1:$GWAY_PORT/ipfs/$FILE_CID?format=$format" > curl_output 2>&1 &&
			ipfs cat $FILE_CID > cat_output &&
			test_cmp cat_output curl_output &&
			test_should_contain "Content-Type: text/plain" headers &&
			test_should_not_contain "Content-Type: application/$format" headers &&
			test_should_not_contain "Content-Type: application/vnd.ipld.dag-$format" headers
			'
			*/
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
			/**
			test_expect_success "GET UnixFS with 'Accept: application/$format' (not dag-$format) is no-op (no conversion)" '
			curl -sD headers -H "Accept: application/$format" "http://127.0.0.1:$GWAY_PORT/ipfs/$FILE_CID" > curl_output 2>&1 &&
			ipfs cat $FILE_CID > cat_output &&
			test_cmp cat_output curl_output &&
			test_should_contain "Content-Type: text/plain" headers &&
			test_should_not_contain "Content-Type: application/$format" headers &&
			test_should_not_contain "Content-Type: application/vnd.ipld.dag-$format" headers
			'
			*/
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
		plain := car.MustOpenUnixfsCar(Fmt("t0123/plain.{{format}}.car", row.Format)).MustGetRoot()
		plainOrDag := car.MustOpenUnixfsCar(Fmt("t0123/plain-that-can-be-dag.{{format}}.car", row.Format)).MustGetRoot()
		formatted := plainOrDag.Formatted("dag-" + row.Format)

		plainCID := plain.Cid()
		plainOrDagCID := plainOrDag.Cid()

		tests := SugarTests{
			/**
			# no explicit format, just codec in CID
			test_expect_success "GET $name without Accept or format= has expected $format Content-Type and body as-is" '
			CID=$(echo "{ \"test\": \"plain json\" }" | ipfs dag put --input-codec json --store-codec $format) &&
			curl -sD headers "http://127.0.0.1:$GWAY_PORT/ipfs/$CID" > curl_output 2>&1 &&
			ipfs block get $CID > ipfs_block_output 2>&1 &&
			test_cmp ipfs_block_output curl_output &&
			test_should_contain "Content-Disposition: ${disposition}\; filename=\"${CID}.${format}\"" headers &&
			test_should_contain "Content-Type: application/$format" headers
			'
			*/
			{
				Name: Fmt(`GET {{name}} without Accept or format= has expected "{{format}}" Content-Type and body as-is`, row.Name, row.Format),
				Hint: `
				No explicit format, just codec in CID
				`,
				Request: Request().
					Path("/ipfs/{{cid}}", plainCID),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Disposition").
							Contains(Fmt(`{{disposition}}; filename="{{cid}}.{{format}}"`, row.Disposition, plainCID, row.Format)),
						Header("Content-Type").
							Contains(Fmt("application/{{format}}", row.Format)),
					).Body(
					plain.RawData(),
				),
			},
			/**
			# explicit format still gives correct output, just codec in CID
			test_expect_success "GET $name with ?format= has expected $format Content-Type and body as-is" '
			CID=$(echo "{ \"test\": \"plain json\" }" | ipfs dag put --input-codec json --store-codec $format) &&
			curl -sD headers "http://127.0.0.1:$GWAY_PORT/ipfs/$CID?format=$format" > curl_output 2>&1 &&
			ipfs block get $CID > ipfs_block_output 2>&1 &&
			test_cmp ipfs_block_output curl_output &&
			test_should_contain "Content-Disposition: ${disposition}\; filename=\"${CID}.${format}\"" headers &&
			test_should_contain "Content-Type: application/$format" headers
			'
			*/
			{
				Name: Fmt("GET {{name}} with ?format= has expected {{format}} Content-Type and body as-is", row.Name, row.Format),
				Hint: `
				Explicit format still gives correct output, just codec in CID
				`,
				Request: Request().
					Path("/ipfs/{{cid}}", plainCID).
					Query("format", row.Format),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Disposition").
							Contains(`{{disposition}}; filename="{{cid}}.{{format}}"`, row.Disposition, plainCID, row.Format),
						Header("Content-Type").
							Contains("application/{{format}}", row.Format),
					).Body(
					plain.RawData(),
				),
			},
			/**
			# explicit format still gives correct output, just codec in CID
			test_expect_success "GET $name with Accept has expected $format Content-Type and body as-is" '
			CID=$(echo "{ \"test\": \"plain json\" }" | ipfs dag put --input-codec json --store-codec $format) &&
			curl -sD headers -H "Accept: application/$format" "http://127.0.0.1:$GWAY_PORT/ipfs/$CID" > curl_output 2>&1 &&
			ipfs block get $CID > ipfs_block_output 2>&1 &&
			test_cmp ipfs_block_output curl_output &&
			test_should_contain "Content-Disposition: ${disposition}\; filename=\"${CID}.${format}\"" headers &&
			test_should_contain "Content-Type: application/$format" headers
			'
			*/
			{
				Name: Fmt("GET {{name}} with Accept has expected {{format}} Content-Type and body as-is", row.Name, row.Format),
				Hint: `
				Explicit format still gives correct output, just codec in CID
				`,
				Request: Request().
					Path("/ipfs/{{cid}}", plainCID).
					Header("Accept", Fmt("application/{{format}}", row.Format)),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Disposition").
							Contains(`{{disposition}}; filename="{{cid}}.{{format}}"`, row.Disposition, plainCID, row.Format),
						Header("Content-Type").
							Contains("application/{{format}}", row.Format),
					).Body(
					plain.RawData(),
				),
			},
			/**
			# explicit dag-* format passed, attempt to parse as dag* variant
			## Note: this works only for simple JSON that can be upgraded to  DAG-JSON.
			test_expect_success "GET $name with format=dag-$format interprets $format as dag-* variant and produces expected Content-Type and body" '
			CID=$(echo "{ \"test\": \"plain-json-that-can-also-be-dag-json\" }" | ipfs dag put --input-codec json --store-codec $format) &&
			curl -sD headers "http://127.0.0.1:$GWAY_PORT/ipfs/$CID?format=dag-$format" > curl_output_param 2>&1 &&
			ipfs dag get --output-codec dag-$format $CID > ipfs_dag_get_output 2>&1 &&
			test_cmp ipfs_dag_get_output curl_output_param &&
			test_should_contain "Content-Disposition: ${disposition}\; filename=\"${CID}.${format}\"" headers &&
			test_should_contain "Content-Type: application/vnd.ipld.dag-$format" headers &&
			curl -s -H "Accept: application/vnd.ipld.dag-$format" "http://127.0.0.1:$GWAY_PORT/ipfs/$CID" > curl_output_accept 2>&1 &&
			test_cmp curl_output_param curl_output_accept
			'
			*/
			{
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
					row.Checker(formatted),
				),
			},
		}

		RunWithSpecs(t, tests, specs.PathGatewayDAG)
	}
}

// ## Pathing, traversal over DAG-JSON and DAG-CBOR
func TestPathing(t *testing.T) {
	dagJSONTraversal := car.MustOpenUnixfsCar("t0123/dag-json-traversal.car").MustGetRoot()
	dagCBORTraversal := car.MustOpenUnixfsCar("t0123/dag-cbor-traversal.car").MustGetRoot()

	dagJSONTraversalCID := dagJSONTraversal.Cid()
	dagCBORTraversalCID := dagCBORTraversal.Cid()

	tests := SugarTests{
		/**
		  test_expect_success "GET DAG-JSON traversal returns 501 if there is path remainder" '
		  curl -sD - "http://127.0.0.1:$GWAY_PORT/ipfs/$DAG_JSON_TRAVERSAL_CID/foo?format=dag-json" > curl_output 2>&1 &&
		  test_should_contain "501 Not Implemented" curl_output &&
		  test_should_contain "reading IPLD Kinds other than Links (CBOR Tag 42) is not implemented" curl_output
		  '
		*/
		{
			Name: "GET DAG-JSON traversal returns 501 if there is path remainder",
			Request: Request().
				Path("/ipfs/{{cid}}/foo", dagJSONTraversalCID).
				Query("format", "dag-json"),
			Response: Expect().
				Status(501).
				Body(Contains("reading IPLD Kinds other than Links (CBOR Tag 42) is not implemented")),
		},
		/**
		  test_expect_success "GET DAG-JSON traverses multiple links" '
		  curl -s "http://127.0.0.1:$GWAY_PORT/ipfs/$DAG_JSON_TRAVERSAL_CID/foo/link/bar?format=dag-json" > curl_output 2>&1 &&
		  jq --sort-keys . curl_output > actual &&
		  echo "{ \"hello\": \"this is not a link\" }" | jq --sort-keys . > expected &&
		  test_cmp expected actual
		  '
		*/
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
		/**
		  test_expect_success "GET DAG-CBOR traversal returns 501 if there is path remainder" '
		  curl -sD - "http://127.0.0.1:$GWAY_PORT/ipfs/$DAG_CBOR_TRAVERSAL_CID/foo?format=dag-cbor" > curl_output 2>&1 &&
		  test_should_contain "501 Not Implemented" curl_output &&
		  test_should_contain "reading IPLD Kinds other than Links (CBOR Tag 42) is not implemented" curl_output
		  '
		*/
		{
			Name: "GET DAG-CBOR traversal returns 501 if there is path remainder",
			Request: Request().
				Path("/ipfs/{{cid}}/foo", dagCBORTraversalCID).
				Query("format", "dag-cbor"),
			Response: Expect().
				Status(501).
				Body(Contains("reading IPLD Kinds other than Links (CBOR Tag 42) is not implemented")),
		},
		/**
		  test_expect_success "GET DAG-CBOR traverses multiple links" '
		  curl -s "http://127.0.0.1:$GWAY_PORT/ipfs/$DAG_CBOR_TRAVERSAL_CID/foo/link/bar?format=dag-json" > curl_output 2>&1 &&
		  jq --sort-keys . curl_output > actual &&
		  echo "{ \"hello\": \"this is not a link\" }" | jq --sort-keys . > expected &&
		  test_cmp expected actual
		  '
		*/
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
	}

	RunWithSpecs(t, tests, specs.PathGatewayDAG)
}

// ## NATIVE TESTS for DAG-JSON (0x0129) and DAG-CBOR (0x71):
// ## DAG- regression tests for core behaviors when native DAG-(CBOR|JSON) is requested
func TestNativeDag(t *testing.T) {
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
		dagTraversal := car.MustOpenUnixfsCar(Fmt("t0123/dag-{{format}}-traversal.car", row.Format)).MustGetRoot()
		dagTraversalCID := dagTraversal.Cid()
		formatted := dagTraversal.Formatted("dag-" + row.Format)

		tests := SugarTests{
			/**
			  # GET without explicit format and Accept: text/html returns raw block

			  test_expect_success "GET $name from /ipfs without explicit format returns the same payload as the raw block" '
			  ipfs block get "/ipfs/$CID" > expected &&
			  curl -sX GET "http://127.0.0.1:$GWAY_PORT/ipfs/$CID" -o curl_output &&
			  test_cmp expected curl_output
			  '
			*/
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
			/**
			  # GET dag-cbor block via Accept and ?format and ensure both are the same as `ipfs block get` output

			  test_expect_success "GET $name from /ipfs with format=dag-$format returns the same payload as the raw block" '
			  ipfs block get "/ipfs/$CID" > expected &&
			  curl -sX GET "http://127.0.0.1:$GWAY_PORT/ipfs/$CID?format=dag-$format" -o curl_ipfs_dag_param_output &&
			  test_cmp expected curl_ipfs_dag_param_output
			  '
			*/
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
			/**
			  test_expect_success "GET $name from /ipfs for application/$format returns the same payload as format=dag-$format" '
			  curl -sX GET "http://127.0.0.1:$GWAY_PORT/ipfs/$CID?format=dag-$format" -o expected &&
			  curl -sX GET "http://127.0.0.1:$GWAY_PORT/ipfs/$CID?format=$format" -o plain_output &&
			  test_cmp expected plain_output
			  '
			  Note: we skip this test since we compare responses bytes to bytes above.
			*/
			/**
			  test_expect_success "GET $name from /ipfs with application/vnd.ipld.dag-$format returns the same payload as the raw block" '
			  ipfs block get "/ipfs/$CID" > expected_block &&
			  curl -sX GET -H "Accept: application/vnd.ipld.dag-$format" "http://127.0.0.1:$GWAY_PORT/ipfs/$CID" -o curl_ipfs_dag_block_accept_output &&
			  test_cmp expected_block curl_ipfs_dag_block_accept_output
			  '
			*/
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
			/**
			  # Make sure DAG-* can be requested as plain JSON or CBOR and response has plain Content-Type for interop purposes

			  test_expect_success "GET $name with format=$format returns same payload as format=dag-$format but with plain Content-Type" '
			  curl -s "http://127.0.0.1:$GWAY_PORT/ipfs/$CID?format=dag-$format" -o expected &&
			  curl -sD plain_headers "http://127.0.0.1:$GWAY_PORT/ipfs/$CID?format=$format" -o plain_output &&
			  test_should_contain "Content-Type: application/$format" plain_headers &&
			  test_cmp expected plain_output
			  '
			*/
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
			/**
			  test_expect_success "GET $name with Accept: application/$format returns same payload as application/vnd.ipld.dag-$format but with plain Content-Type" '
			  curl -s -H "Accept: application/vnd.ipld.dag-$format" "http://127.0.0.1:$GWAY_PORT/ipfs/$CID" > expected &&
			  curl -sD plain_headers -H "Accept: application/$format" "http://127.0.0.1:$GWAY_PORT/ipfs/$CID" > plain_output &&
			  test_should_contain "Content-Type: application/$format" plain_headers &&
			  test_cmp expected plain_output
			  '
			*/
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
			/**
			  # Make sure expected HTTP headers are returned with the dag- block

			  test_expect_success "GET response for application/vnd.ipld.dag-$format has expected Content-Type" '
			  curl -svX GET -H "Accept: application/vnd.ipld.dag-$format" "http://127.0.0.1:$GWAY_PORT/ipfs/$CID" >/dev/null 2>curl_output &&
			  test_should_contain "< Content-Type: application/vnd.ipld.dag-$format" curl_output
			  '
			  test_expect_success "GET response for application/vnd.ipld.dag-$format includes Content-Length" '
			  BYTES=$(ipfs block get $CID | wc --bytes)
			  test_should_contain "< Content-Length: $BYTES" curl_output
			  '
			  test_expect_success "GET response for application/vnd.ipld.dag-$format includes Content-Disposition" '
			  test_should_contain "< Content-Disposition: ${disposition}\; filename=\"${CID}.${format}\"" curl_output
			  '
			  test_expect_success "GET response for application/vnd.ipld.dag-$format includes nosniff hint" '
			  test_should_contain "< X-Content-Type-Options: nosniff" curl_output
			  '
			*/
			{
				Name: Fmt("GET response for application/vnd.ipld.dag-{{format}} has expected Content-Type", row.Format),
				Hint: `Make sure expected HTTP headers are returned with the dag- block`,
				Request: Request().
					Path("/ipfs/{{cid}}", dagTraversalCID).
					Header("Accept", Fmt("application/vnd.ipld.dag-{{format}}", row.Format)),
				Response: Expect().
					Headers(
						Header("Content-Type").Hint("expected Content-Type").Equals("application/vnd.ipld.dag-{{format}}", row.Format),
						Header("Content-Length").Hint("includes Content-Length").Equals("{{length}}", len(dagTraversal.RawData())),
						Header("Content-Disposition").Hint("includes Content-Disposition").Contains(`{{disposition}}; filename="{{cid}}.{{format}}"`, row.Disposition, dagTraversalCID, row.Format),
						Header("X-Content-Type-Options").Hint("includes nosniff hint").Contains("nosniff"),
					),
			},
			/**
			  test_expect_success "GET for application/vnd.ipld.dag-$format with query filename includes Content-Disposition with custom filename" '
			  curl -svX GET -H "Accept: application/vnd.ipld.dag-$format" "http://127.0.0.1:$GWAY_PORT/ipfs/$CID?filename=foobar.$format" >/dev/null 2>curl_output_filename &&
			  test_should_contain "< Content-Disposition: ${disposition}\; filename=\"foobar.$format\"" curl_output_filename
			  '
			*/
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
			/**
			  test_expect_success "GET for application/vnd.ipld.dag-$format with ?download=true forces Content-Disposition: attachment" '
			  curl -svX GET -H "Accept: application/vnd.ipld.dag-$format" "http://127.0.0.1:$GWAY_PORT/ipfs/$CID?filename=foobar.$format&download=true" >/dev/null 2>curl_output_filename &&
			  test_should_contain "< Content-Disposition: attachment" curl_output_filename
			  '
			*/
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
			/**
			  # Cache control HTTP headers
			  # (basic checks, detailed behavior is tested in  t0116-gateway-cache.sh)

			  test_expect_success "GET response for application/vnd.ipld.dag-$format includes Etag" '
			  test_should_contain "< Etag: \"${CID}.dag-$format\"" curl_output
			  '
			  test_expect_success "GET response for application/vnd.ipld.dag-$format includes X-Ipfs-Path and X-Ipfs-Roots" '
			  test_should_contain "< X-Ipfs-Path" curl_output &&
			  test_should_contain "< X-Ipfs-Roots" curl_output
			  '
			  test_expect_success "GET response for application/vnd.ipld.dag-$format includes Cache-Control" '
			  test_should_contain "< Cache-Control: public, max-age=29030400, immutable" curl_output
			  '
			*/
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
			/**
			  # HTTP HEAD behavior
			  test_expect_success "HEAD $name with no explicit format returns HTTP 200" '
			  curl -I "http://127.0.0.1:$GWAY_PORT/ipfs/$CID" -o output &&
			  test_should_contain "HTTP/1.1 200 OK" output &&
			  test_should_contain "Content-Type: application/vnd.ipld.dag-$format" output &&
			  test_should_contain "Content-Length: " output
			  '
			*/
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
			/**
			  test_expect_success "HEAD $name with an explicit DAG-JSON format returns HTTP 200" '
			  curl -I "http://127.0.0.1:$GWAY_PORT/ipfs/$CID?format=dag-json" -o output &&
			  test_should_contain "HTTP/1.1 200 OK" output &&
			  test_should_contain "Etag: \"$CID.dag-json\"" output &&
			  test_should_contain "Content-Type: application/vnd.ipld.dag-json" output &&
			  test_should_contain "Content-Length: " output
			  '
			*/
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
			/**
			  test_expect_success "HEAD $name with only-if-cached for missing block returns HTTP 412 Precondition Failed" '
			  MISSING_CID=$(echo "{\"t\": \"$(date +{{}})\"}" | ipfs dag put --store-codec=dag-${format}) &&
			  ipfs block rm -f -q $MISSING_CID &&
			  curl -I -H "Cache-Control: only-if-cached" "http://127.0.0.1:$GWAY_PORT/ipfs/$MISSING_CID" -o output &&
			  test_should_contain "HTTP/1.1 412 Precondition Failed" output
			  '
			*/
			{
				Name: Fmt("HEAD {{name}} with only-if-cached for missing block returns HTTP 412 Precondition Failed", row.Name),
				Request: Request().
					Path("/ipfs/{{cid}}", missingCID).
					Header("Cache-Control", "only-if-cached").
					Method("HEAD"),
				Response: Expect().
					Status(412),
			},
			// test_expect_success "GET $name on /ipfs with Accept: text/html returns HTML (dag-index-html)" '
			// curl -sD - -H "Accept: text/html" "http://127.0.0.1:$GWAY_PORT/ipfs/$CID" > curl_output 2>&1 &&
			// test_should_not_contain "Content-Disposition" curl_output &&
			// test_should_not_contain "Cache-Control" curl_output &&
			// test_should_contain "Etag: \"DagIndex-" curl_output &&
			// test_should_contain "Content-Type: text/html" curl_output &&
			// test_should_contain "</html>" curl_output
			// '
			{
				Name: Fmt("GET {{name}} on /ipfs with Accept: text/html returns HTML (dag-index-html)", row.Name),
				Request: Request().
					Path("/ipfs/{{cid}}/", dagTraversalCID).
					Header("Accept", "text/html"),
				Response: Expect().
					Headers(
						Header("Etag").Contains("DagIndex-"),
						Header("Content-Type").Contains("text/html"),
						Header("Content-Disposition").IsEmpty(),
						Header("Cache-Control").IsEmpty(),
					).Body(
					Contains("</html>"),
				),
			},
		}

		RunWithSpecs(t, tests, specs.PathGatewayDAG)
	}
}

func TestGatewayJSONCborAndIPNS(t *testing.T) {
	ipnsIdDagJSON := "k51qzi5uqu5dhjghbwdvbo6mi40htrq6e2z4pwgp15pgv3ho1azvidttzh8yy2"
	ipnsIdDagCBOR := "k51qzi5uqu5dghjous0agrwavl8vzl64xckoqzwqeqwudfr74kfd11zcyk3b7l"

	ipnsDagJSON := ipns.MustOpenIPNSRecordWithKey(Fmt("t0123/{{id}}.ipns-record", ipnsIdDagJSON))
	ipnsDagCBOR := ipns.MustOpenIPNSRecordWithKey(Fmt("t0123/{{id}}.ipns-record", ipnsIdDagCBOR))

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
		plain := car.MustOpenUnixfsCar(Fmt("t0123/dag-{{format}}-traversal.car", row.Format)).MustGetRoot()
		plainCID := plain.Cid()

		// # IPNS behavior (should be same as immutable /ipfs, but with different caching headers)
		// # To keep tests small we only confirm payload is the same, and then only test delta around caching headers.
		tests = append(tests, SugarTests{
			// test_expect_success "GET $name from /ipns without explicit format returns the same payload as /ipfs" '
			// curl -sX GET "http://127.0.0.1:$GWAY_PORT/ipfs/$CID" -o ipfs_output &&
			// curl -sX GET "http://127.0.0.1:$GWAY_PORT/ipns/$IPNS_ID" -o ipns_output &&
			// test_cmp ipfs_output ipns_output
			// '
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
			// test_expect_success "GET $name from /ipns without explicit format returns the same payload as /ipfs" '
			// curl -sX GET "http://127.0.0.1:$GWAY_PORT/ipfs/$CID?format=dag-$format" -o ipfs_output &&
			// curl -sX GET "http://127.0.0.1:$GWAY_PORT/ipns/$IPNS_ID?format=dag-$format" -o ipns_output &&
			// test_cmp ipfs_output ipns_output
			// '
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

			// test_expect_success "GET $name from /ipns with explicit application/vnd.ipld.dag-$format has expected headers" '
			// curl -svX GET -H "Accept: application/vnd.ipld.dag-$format" "http://127.0.0.1:$GWAY_PORT/ipns/$IPNS_ID" >/dev/null 2>curl_output &&
			// test_should_not_contain "Cache-Control" curl_output &&
			// test_should_contain "< Content-Type: application/vnd.ipld.dag-$format" curl_output &&
			// test_should_contain "< Etag: \"${CID}.dag-$format\"" curl_output &&
			// test_should_contain "< X-Ipfs-Path" curl_output &&
			// test_should_contain "< X-Ipfs-Roots" curl_output
			// '
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
			// # When Accept header includes text/html and no explicit format is requested for DAG-(CBOR|JSON)
			// # The gateway returns generated HTML index (see dag-index-html) for web browsers (similar to dir-index-html returned for unixfs dirs)
			// # As this is generated, we don't return immutable Cache-Control, even on /ipfs (same as for dir-index-html)

			// test_expect_success "GET $name on /ipns with Accept: text/html returns HTML (dag-index-html)" '
			// curl -sD - -H "Accept: text/html" "http://127.0.0.1:$GWAY_PORT/ipns/$IPNS_ID" > curl_output 2>&1 &&
			// test_should_not_contain "Content-Disposition" curl_output &&
			// test_should_not_contain "Cache-Control" curl_output &&
			// test_should_contain "Etag: \"DagIndex-" curl_output &&
			// test_should_contain "Content-Type: text/html" curl_output &&
			// test_should_contain "</html>" curl_output
			// '
			{
				Name: Fmt("GET {{name}} on /ipns with Accept: text/html returns HTML (dag-index-html)", row.Name),
				Request: Request().
					Path("/ipns/{{id}}/", row.fixture.Key()).
					Header("Accept", "text/html"),
				Response: Expect().
					Headers(
						Header("Etag").Contains("DagIndex-"),
						Header("Content-Type").Contains("text/html"),
						Header("Content-Disposition").IsEmpty(),
						Header("Cache-Control").IsEmpty(),
					).Body(
					Contains("</html>"),
				),
			},
		}...)
	}

	RunWithSpecs(t, tests, specs.PathGatewayDAG, specs.PathGatewayIPNS)
}
