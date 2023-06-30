package tests

import (
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
	"github.com/ipfs/gateway-conformance/tooling/tmpl"
)

func TestTar(t *testing.T) {
	fixtureOutside := car.MustOpenUnixfsCar("path_gateway_tar/outside-root.car")
	fixtureInside := car.MustOpenUnixfsCar("path_gateway_tar/inside-root.car")

	outsideRootCID := fixtureOutside.MustGetCid()
	insideRootCID := fixtureInside.MustGetCid()

	fixture := car.MustOpenUnixfsCar("path_gateway_tar/fixtures.car")
	dirCID := fixture.MustGetCid() // root dir
	fileCID := fixture.MustGetCid("ą", "ę", "file-źł.txt")

	tests := SugarTests{
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
		{
			Name: "GET TAR with explicit ?filename= succeeds with modified Content-Disposition header",
			Request: Request().
				Path("/ipfs/{{cid}}", dirCID).
				Query("filename", "testтест.tar").
				Query("format", "tar"),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Disposition").Contains(`attachment; filename="test____.tar"; filename*=UTF-8''test%D1%82%D0%B5%D1%81%D1%82.tar`),
				),
		},
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
