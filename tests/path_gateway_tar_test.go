package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
	"github.com/ipfs/gateway-conformance/tooling/tmpl"
)

// TODO(laurent): this was t0122_gateway_tar_test.go

// "Test HTTP Gateway TAR (application/x-tar) Support"
func TestTar(t *testing.T) {
	/**
	test_expect_success "Add CARs with relative paths to test with" '
	ipfs dag import ../t0122-gateway-tar/outside-root.car > import_output &&
	test_should_contain $OUTSIDE_ROOT_CID import_output &&
	ipfs dag import ../t0122-gateway-tar/inside-root.car > import_output &&
	test_should_contain $INSIDE_ROOT_CID import_output
	'
	*/
	fixtureOutside := car.MustOpenUnixfsCar("t0122/outside-root.car")
	fixtureInside := car.MustOpenUnixfsCar("t0122/inside-root.car")

	outsideRootCID := fixtureOutside.MustGetCid()
	insideRootCID := fixtureInside.MustGetCid()

	// OUTSIDE_ROOT_CID="bafybeicaj7kvxpcv4neaqzwhrqqmdstu4dhrwfpknrgebq6nzcecfucvyu"
	// INSIDE_ROOT_CID="bafybeibfevfxlvxp5vxobr5oapczpf7resxnleb7tkqmdorc4gl5cdva3y"
	//
	// test_expect_success "Add the test directory" '
	//   ipfs dag import ../t0122-gateway-tar/fixtures.car
	// '
	// DIR_CID=bafybeig6ka5mlwkl4subqhaiatalkcleo4jgnr3hqwvpmsqfca27cijp3i # ./rootDir
	// FILE_CID=bafkreialihlqnf5uwo4byh4n3cmwlntwqzxxs2fg5vanqdi3d7tb2l5xkm # ./rootDir/ą/ę/file-źł.txt
	// FILE_SIZE=34

	fixture := car.MustOpenUnixfsCar("t0122/fixtures.car")
	dirCID := fixture.MustGetCid() // root dir
	fileCID := fixture.MustGetCid("ą", "ę", "file-źł.txt")

	tests := SugarTests{
		/**
		  test_expect_success "GET TAR with format=tar and extract" '
		  curl "http://127.0.0.1:$GWAY_PORT/ipfs/$FILE_CID?format=tar" | tar -x
		  '
		  test_expect_success "GET TAR with format=tar has expected Content-Type" '
		  curl -sD - "http://127.0.0.1:$GWAY_PORT/ipfs/$FILE_CID?format=tar" > curl_output_filename 2>&1 &&
		  test_should_contain "Content-Disposition: attachment;" curl_output_filename &&
		  test_should_contain "Etag: W/\"$FILE_CID.x-tar" curl_output_filename &&
		  test_should_contain "Content-Type: application/x-tar" curl_output_filename
		  '
		  test_expect_success "GET TAR has expected root file" '
		  rm -rf outputDir && mkdir outputDir &&
		  curl "http://127.0.0.1:$GWAY_PORT/ipfs/$FILE_CID?format=tar" | tar -x -C outputDir &&
		  test -f "outputDir/$FILE_CID" &&
		  echo "I am a txt file on path with utf8" > expected &&
		  test_cmp expected outputDir/$FILE_CID
		  '
		*/
		{
			Name: "GET TAR with format=tar and extract",
			Request: Request().
				Path("/ipfs/{{cid}}", fileCID).
				Query("format", "tar"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Disposition").Contains("attachment;"),
					Header("Etag").Contains(`W/"{{cid}}.x-tar`, fileCID),
					Header("Content-Type").Contains("application/x-tar"),
				).Body(
				IsTarFile().
					HasFileWithContent(
						fileCID,
						"I am a txt file on path with utf8\n",
					),
			),
		},
		/**
		  test_expect_success "GET TAR with 'Accept: application/x-tar' and extract" '
		  curl -H "Accept: application/x-tar" "http://127.0.0.1:$GWAY_PORT/ipfs/$FILE_CID" | tar -x
		  '
		  test_expect_success "GET TAR with 'Accept: application/x-tar' has expected Content-Type" '
		  curl -sD - -H "Accept: application/x-tar" "http://127.0.0.1:$GWAY_PORT/ipfs/$FILE_CID" > curl_output_filename 2>&1 &&
		  test_should_contain "Content-Disposition: attachment;" curl_output_filename &&
		  test_should_contain "Etag: W/\"$FILE_CID.x-tar" curl_output_filename &&
		  test_should_contain "Content-Type: application/x-tar" curl_output_filename
		  '
		*/
		{
			Name: "GET TAR with 'Accept: application/x-tar' and extract",
			Request: Request().
				Path("/ipfs/{{cid}}", fileCID).
				Header("Accept", "application/x-tar"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Disposition").Contains("attachment;"),
					Header("Etag").Contains(`W/"{{cid}}.x-tar`, fileCID),
					Header("Content-Type").Contains("application/x-tar"),
				).Body(
				IsTarFile(),
			),
		},
		/**
		  test_expect_success "GET TAR has expected root directory" '
		  rm -rf outputDir && mkdir outputDir &&
		  curl "http://127.0.0.1:$GWAY_PORT/ipfs/$DIR_CID?format=tar" | tar -x -C outputDir &&
		  test -d "outputDir/$DIR_CID" &&
		  echo "I am a txt file on path with utf8" > expected &&
		  test_cmp expected outputDir/$DIR_CID/ą/ę/file-źł.txt
		  '
		*/
		{
			Name: "GET TAR has expected root directory",
			Request: Request().
				Path("/ipfs/{{cid}}", dirCID).
				Query("format", "tar"),
			Response: Expect().
				Status(200).
				Body(
					IsTarFile().
						HasFileWithContent(
							tmpl.Fmt("{{cid}}/ą/ę/file-źł.txt", dirCID),
							"I am a txt file on path with utf8\n",
						),
				),
		},
		/**
		  test_expect_success "GET TAR with explicit ?filename= succeeds with modified Content-Disposition header" "
		  curl -fo actual -D actual_headers 'http://127.0.0.1:$GWAY_PORT/ipfs/$DIR_CID?filename=testтест.tar&format=tar' &&
		  grep -F 'Content-Disposition: attachment; filename=\"test____.tar\"; filename*=UTF-8'\'\''test%D1%82%D0%B5%D1%81%D1%82.tar' actual_headers
		  "
		*/
		{
			Name: "GET TAR with explicit ?filename= succeeds with modified Content-Disposition header",
			Request: Request().
				Path("/ipfs/{{cid}}", dirCID).
				Query("filename", "testтест.tar").
				Query("format", "tar"),
			Response: Expect().
				Status(200).
				Headers(
					// Note: golang's compiler assumes the "weird" string is a format string, we use `"{{x}}"` to move it out of the way.
					Header("Content-Disposition").Contains("{{x}}", `attachment; filename="test____.tar"; filename*=UTF-8''test%D1%82%D0%B5%D1%81%D1%82.tar`),
				),
		},
		/**
		  test_expect_success "GET TAR with relative paths outside root fails" '
		  curl -o - "http://127.0.0.1:$GWAY_PORT/ipfs/$OUTSIDE_ROOT_CID?format=tar" > curl_output_filename &&
		  test_should_contain "relative UnixFS paths outside the root are now allowed" curl_output_filename
		  '
		*/
		{
			Name: "GET TAR with relative paths outside root fails",
			Request: Request().
				Path("/ipfs/{{cid}}", outsideRootCID).
				Query("format", "tar"),
			Response: Expect().
				Body(
					Contains("relative UnixFS paths outside the root are now allowed"),
				),
		},
		/**
		  test_expect_success "GET TAR with relative paths inside root works" '
		  rm -rf outputDir && mkdir outputDir &&
		  curl "http://127.0.0.1:$GWAY_PORT/ipfs/$INSIDE_ROOT_CID?format=tar" | tar -x -C outputDir &&
		  test -f outputDir/$INSIDE_ROOT_CID/foobar/file
		  '
		*/
		{
			Name: "GET TAR with relative paths inside root works",
			Request: Request().
				Path("/ipfs/{{cid}}", insideRootCID).
				Query("format", "tar"),
			Response: Expect().
				Status(200).
				Body(
					IsTarFile().
						HasFile(
							"{{cid}}/foobar/file", insideRootCID,
						),
				),
		},
	}

	RunWithSpecs(t, tests, specs.PathGatewayTAR)
}
