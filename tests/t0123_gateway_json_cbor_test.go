package tests

import (
	"fmt"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/test"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewayJsonCbor(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0123-gateway-json-cbor.car")

	dirCID := fixture.MustGetCid() // root dir
	fileJSONCID := fixture.MustGetCid("ą", "ę", "t.json")
	fileJSONData := fixture.MustGetRawData("ą", "ę", "t.json")
	fileCID := fixture.MustGetCid("ą", "ę", "file-źł.txt")
	fileSize := len(fixture.MustGetRawData("ą", "ę", "file-źł.txt"))

	fmt.Println("rootDirCID:", dirCID)
	fmt.Println("fileJSONCID:", fileJSONCID)
	fmt.Println("fileJSONData:", fileJSONData)
	fmt.Println("fileCID:", fileCID)
	fmt.Println("fileSize:", fileSize)

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
				Path("ipfs/%s", fileJSONCID).
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
				Path("ipfs/%s", fileJSONCID).
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

	test.Run(t, tests.Build())
}

// ## Reading UnixFS (data encoded with dag-pb codec) as DAG-CBOR and DAG-JSON
// ## (returns representation defined in https://ipld.io/specs/codecs/dag-pb/spec/#logical-format)
func TestDAgPbConversion(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0123-gateway-json-cbor.car")

	dirCID := fixture.MustGetCid() // root dir
	fileCID := fixture.MustGetCid("ą", "ę", "file-źł.txt")
	fileData := fixture.MustGetRawData("ą", "ę", "file-źł.txt")

	table := []struct {
		Name        string
		Format      string
		Disposition string
	}{
		{"DAG-JSON", "json", "inline"},
		{"DAG-CBOR", "cbor", "attachment"},
	}

	for _, row := range table {
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
				Name: fmt.Sprintf("GET UnixFS file as %s with format=dag-%s converts to the expected Content-Type", row.Name, row.Format),
				Request: Request().
					Path("ipfs/%s?format=dag-%s", fileCID, row.Format),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Type").
							Equals("application/vnd.ipld.dag-%s", row.Format),
						Header("Content-Disposition").
							Contains("%s; filename=\"%s.%s\"", row.Disposition, fileCID, row.Format),
						Header("Content-Type").
							Not().Contains("application/%s", row.Format),
					),
				// TODO: test body `ipfs dag get --output-codec dag-$format $FILE_CID > ipfs_dag_get_output`
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
				Name: fmt.Sprintf("GET UnixFS directory as %s with format=dag-%s converts to the expected Content-Type", row.Name, row.Format),
				Request: Request().
					Path("ipfs/%s?format=dag-%s", dirCID, row.Format),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Type").
							Equals("application/vnd.ipld.dag-%s", row.Format),
						Header("Content-Disposition").
							Contains("%s; filename=\"%s.%s\"", row.Disposition, dirCID, row.Format),
						Header("Content-Type").
							Not().Contains("application/%s", row.Format),
					),
				// TODO: test body `ipfs dag get --output-codec dag-$format $DIR_CID > ipfs_dag_get_output`
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
				Name: fmt.Sprintf("GET UnixFS as %s with 'Accept: application/vnd.ipld.dag-%s' converts to the expected Content-Type", row.Name, row.Format),
				Request: Request().
					Path("ipfs/%s", fileCID).
					Headers(
						Header("Accept", "application/vnd.ipld.dag-%s", row.Format),
					),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Disposition").
							Contains("%s; filename=\"%s.%s\"", row.Disposition, fileCID, row.Format),
						Header("Content-Type").
							Equals("application/vnd.ipld.dag-%s", row.Format),
						Header("Content-Type").
							Not().Contains("application/%s", row.Format),
					),
			},
			/**
			test_expect_success "GET UnixFS as $name with 'Accept: foo, application/vnd.ipld.dag-$format,bar' converts to the expected Content-Type" '
			curl -sD - -H "Accept: foo, application/vnd.ipld.dag-$format,text/plain" "http://127.0.0.1:$GWAY_PORT/ipfs/$FILE_CID" > curl_output 2>&1 &&
			test_should_contain "Content-Type: application/vnd.ipld.dag-$format" curl_output
			'
			*/
			{
				Name: fmt.Sprintf("GET UnixFS as %s with 'Accept: foo, application/vnd.ipld.dag-%s,bar' converts to the expected Content-Type", row.Name, row.Format),
				Request: Request().
					Path("ipfs/%s", fileCID).
					Headers(
						Header("Accept", "foo, application/vnd.ipld.dag-%s,bar", row.Format),
					),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Type").
							Equals("application/vnd.ipld.dag-%s", row.Format),
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
				Name: fmt.Sprintf("GET UnixFS with format=%s (not dag-%s) is no-op (no conversion)", row.Format, row.Format),
				Request: Request().
					Path("ipfs/%s?format=%s", fileCID, row.Format),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Type").
							Equals("text/plain"),
						Header("Content-Type").
							Not().Contains("application/%s", row.Format),
						Header("Content-Type").
							Not().Contains("application/vnd.ipld.dag-%s", row.Format),
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
				Name: fmt.Sprintf("GET UnixFS with 'Accept: application/%s' (not dag-%s) is no-op (no conversion)", row.Format, row.Format),
				Request: Request().
					Path("ipfs/%s", fileCID).
					Headers(
						Header("Accept", "application/%s", row.Format),
					),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Type").
							Equals("text/plain"),
						Header("Content-Type").
							Not().Contains("application/%s", row.Format),
						Header("Content-Type").
							Not().Contains("application/vnd.ipld.dag-%s", row.Format),
					).Body(
					fileData,
				),
			},
		}

		test.Run(t, tests.Build())
	}

}

// # Requesting CID with plain json (0x0200) and cbor (0x51) codecs
// # (note these are not UnixFS, not DAG-* variants, just raw block identified by a CID with a special codec)
func TestPlainCodec(t *testing.T) {
	table := []struct {
		Name        string
		Format      string
		Disposition string
	}{
		{"plain JSON codec", "json", "inline"},
		{"plain CBOR codec", "cbor", "attachment"},
	}

	for _, row := range table {
		plainFixture := car.MustOpenRawBlockFromCar(fmt.Sprintf("t0123/plain.%s.car", row.Format))
		plainOrDagFixture := car.MustOpenRawBlockFromCar(fmt.Sprintf("t0123/plain-that-can-be-dag.%s.car", row.Format))

		plainCID := plainFixture.Cid()
		plainOrDagCID := plainOrDagFixture.Cid()

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
				Name: fmt.Sprintf(`GET %s without Accept or format= has expected "%s" Content-Type and body as-is`, row.Name, row.Format),
				Hint: `
				No explicit format, just codec in CID
				`,
				Request: Request().
					Path("ipfs/%s", plainCID),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Disposition").
							Contains(fmt.Sprintf("%s; filename=\"%s.%s\"", row.Disposition, plainCID, row.Format)),
						Header("Content-Type").
							Contains(fmt.Sprintf("application/%s", row.Format)),
					).Body(
					plainFixture.RawData(),
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
				Name: fmt.Sprintf("GET %s with ?format= has expected %s Content-Type and body as-is", row.Name, row.Format),
				Hint: `
				Explicit format still gives correct output, just codec in CID
				`,
				Request: Request().
					Path("ipfs/%s", plainCID).
					Query("format", row.Format),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Disposition").
							Contains("%s; filename=\"%s.%s\"", row.Disposition, plainCID, row.Format),
						Header("Content-Type").
							Contains("application/%s", row.Format),
					).Body(
					plainFixture.RawData(),
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
				Name: fmt.Sprintf("GET %s with Accept has expected %s Content-Type and body as-is", row.Name, row.Format),
				Hint: `
				Explicit format still gives correct output, just codec in CID
				`,
				Request: Request().
					Path("ipfs/%s", plainCID).
					Header("Accept", fmt.Sprintf("application/%s", row.Format)),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Disposition").
							Contains("%s; filename=\"%s.%s\"", row.Disposition, plainCID, row.Format),
						Header("Content-Type").
							Contains("application/%s", row.Format),
					).Body(
					plainFixture.RawData(),
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
				Name: fmt.Sprintf("GET %s with format=dag-%s interprets %s as dag-* variant and produces expected Content-Type and body", row.Name, row.Format, row.Format),
				Hint: `
				Explicit dag-* format passed, attempt to parse as dag* variant
				Note: this works only for simple JSON that can be upgraded to  DAG-JSON.
				`,
				Request: Request().
					Path("ipfs/%s", plainOrDagCID).
					Query("format", fmt.Sprintf("dag-%s", row.Format)),
				Response: Expect().
					Status(200).
					Headers(
						Header("Content-Disposition").
							Contains("%s; filename=\"%s.%s\"", row.Disposition, plainOrDagCID, row.Format),
						Header("Content-Type").
							Contains("application/vnd.ipld.dag-%s", row.Format),
					).Body(
					IsJSONEqual(plainOrDagFixture.RawData()),
				),
			},
		}

		test.Run(t, tests.Build())
	}
}
