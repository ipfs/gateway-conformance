package tests

import (
	"fmt"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
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
