package tests

import (
	"strconv"
	"strings"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling"
	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestTrustlessRaw(t *testing.T) {
	tooling.LogTestGroup(t, GroupTrustlessGateway)

	fixture := car.MustOpenUnixfsCar("gateway-raw-block.car")

	tests := SugarTests{
		{
			Name: "GET with format=raw param returns a raw block",
			Request: Request().
				Path("/ipfs/{{cid}}", fixture.MustGetCid("dir")).
				Query("format", "raw"),
			Response: Expect().
				Status(200).
				Body(fixture.MustGetRawData("dir")),
		},
		{
			Name: "GET with application/vnd.ipld.raw header returns a raw block",
			Request: Request().
				Path("/ipfs/{{cid}}", fixture.MustGetCid("dir")).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
				),
			Response: Expect().
				Status(200).
				Body(fixture.MustGetRawData("dir")),
		},
		{
			Name: "GET with application/vnd.ipld.raw header returns expected response headers",
			Request: Request().
				Path("/ipfs/{{cid}}", fixture.MustGetCid("dir", "ascii.txt")).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
				),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Type").
						Equals("application/vnd.ipld.raw"),
					Header("Content-Length").
						Equals("{{length}}", len(fixture.MustGetRawData("dir", "ascii.txt"))),
					Header("Content-Disposition").
						Contains("attachment;"),
					Header("X-Content-Type-Options").
						Equals("nosniff"),
				).
				Body(fixture.MustGetRawData("dir", "ascii.txt")),
		},
		{
			Name: "GET with application/vnd.ipld.raw header and filename param returns expected Content-Disposition header with custom filename",
			Request: Request().
				Path("/ipfs/{{cid}}?filename=foobar.bin", fixture.MustGetCid("dir", "ascii.txt")).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
				),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Disposition").
						Contains(`attachment;`),
					Header("Content-Disposition").
						Contains(`filename="foobar.bin`),
				),
		},
		{
			Name: "GET with application/vnd.ipld.raw header returns expected caching headers",
			Request: Request().
				Path("/ipfs/{{cid}}", fixture.MustGetCid("dir", "ascii.txt")).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
				),
			Response: Expect().
				Status(200).
				Headers(
					Header("Etag").
						Exists(),
					Header("X-IPFS-Path").
						Equals("/ipfs/{{cid}}", fixture.MustGetCid("dir", "ascii.txt")),
					Header("X-IPFS-Roots").
						Contains(fixture.MustGetCid("dir", "ascii.txt")),
					Header("Cache-Control").
						Hint("It should be public, immutable and have max-age of at least 29030400.").
						Checks(func(v string) bool {
							directives := strings.Split(strings.ReplaceAll(v, " ", ""), ",")
							dir := make(map[string]string)
							for _, directive := range directives {
								parts := strings.Split(directive, "=")
								if len(parts) == 2 {
									dir[parts[0]] = parts[1]
								} else {
									dir[parts[0]] = ""
								}
							}
							if _, ok := dir["public"]; !ok {
								return false
							}
							if _, ok := dir["immutable"]; !ok {
								return false
							}
							if maxAge, ok := dir["max-age"]; ok {
								maxAgeInt, err := strconv.Atoi(maxAge)
								if err != nil {
									return false
								}
								if maxAgeInt < 29030400 {
									return false
								}
							} else {
								return false
							}
							return true
						}),
				),
		},
	}

	RunWithSpecs(t, tests, specs.TrustlessGatewayRaw)
}

func TestTrustlessRawRanges(t *testing.T) {
	tooling.LogTestGroup(t, GroupTrustlessGateway)

	// Multi-range requests MUST conform to the HTTP semantics. The server does not
	// need to be able to support returning multiple ranges. However, it must respond
	// correctly.
	fixture := car.MustOpenUnixfsCar("gateway-raw-block.car")

	var (
		contentType  string
		contentRange string
	)

	RunWithSpecs(t, SugarTests{
		{
			Name: "GETaa with application/vnd.ipld.raw with single range request includes correct bytes",
			Request: Request().
				Path("/ipfs/{{cid}}", fixture.MustGetCid("dir", "ascii.txt")).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
					Header("Range", "bytes=6-16"),
				),
			Response: Expect().
				Status(206).
				Headers(
					Header("Content-Type").Contains("application/vnd.ipld.raw"),
					Header("Content-Range").Equals("bytes 6-16/31"),
				).
				Body(fixture.MustGetRawData("dir", "ascii.txt")[6:17]),
		},
		{
			Name: "GET with application/vnd.ipld.raw with multiple range request includes correct bytes",
			Request: Request().
				Path("/ipfs/{{cid}}", fixture.MustGetCid("dir", "ascii.txt")).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
					Header("Range", "bytes=6-16,0-4"),
				),
			Response: Expect().
				Status(206).
				Headers(
					Header("Content-Type").
						Checks(func(v string) bool {
							contentType = v
							return v != ""
						}),
					Header("Content-Range").
						ChecksAll(func(v []string) bool {
							if len(v) == 1 {
								contentRange = v[0]
							}
							return true
						}),
				),
		},
	}, specs.PathGatewayRaw)

	tests := SugarTests{}

	if strings.Contains(contentType, "application/vnd.ipld.raw") {
		// The server is not able to respond to a multi-range request. Therefore,
		// there might be only one range or... just the whole file, depending on the headers.

		if contentRange == "" {
			// Server does not support range requests and must send back the complete file.
			tests = append(tests, SugarTest{
				Name: "GET with application/vnd.ipld.raw with multiple range request includes correct bytes",
				Request: Request().
					Path("/ipfs/{{cid}}", fixture.MustGetCid("dir", "ascii.txt")).
					Headers(
						Header("Accept", "application/vnd.ipld.raw"),
						Header("Range", "bytes=6-16,0-4"),
					),
				Response: Expect().
					Status(206).
					Headers(
						Header("Content-Type").Contains("application/vnd.ipld.raw"),
						Header("Content-Range").IsEmpty(),
					).
					Body(fixture.MustGetRawData("dir", "ascii.txt")),
			})
		} else {
			// Server supports range requests but only the first range.
			tests = append(tests, SugarTest{
				Name: "GET with application/vnd.ipld.raw with multiple range request includes correct bytes",
				Request: Request().
					Path("/ipfs/{{cid}}", fixture.MustGetCid("dir", "ascii.txt")).
					Headers(
						Header("Accept", "application/vnd.ipld.raw"),
						Header("Range", "bytes=6-16,0-4"),
					),
				Response: Expect().
					Status(206).
					Headers(
						Header("Content-Type").Contains("application/vnd.ipld.raw"),
						Header("Content-Range", "bytes 6-16/31"),
					).
					Body(fixture.MustGetRawData("dir", "ascii.txt")[6:17]),
			})
		}
	} else if strings.Contains(contentType, "multipart/byteranges") {
		// The server supports responding with multi-range requests.
		tests = append(tests, SugarTest{
			Name: "GET with application/vnd.ipld.raw with multiple range request includes correct bytes",
			Request: Request().
				Path("/ipfs/{{cid}}", fixture.MustGetCid("dir", "ascii.txt")).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
					Header("Range", "bytes=6-16,0-4"),
				),
			Response: Expect().
				Status(206).
				Headers(
					Header("Content-Type").Contains("multipart/byteranges"),
				).
				Body(And(
					Contains("Content-Range: bytes 6-16/31"),
					Contains("Content-Type: application/vnd.ipld.raw"),
					Contains(string(fixture.MustGetRawData("dir", "ascii.txt")[6:17])),
					Contains("Content-Range: bytes 0-4/31"),
					Contains(string(fixture.MustGetRawData("dir", "ascii.txt")[0:5])),
				)),
		})
	} else {
		t.Error("Content-Type header did not match any of the accepted options")
	}

	RunWithSpecs(t, tests, specs.TrustlessGatewayRaw)
}
