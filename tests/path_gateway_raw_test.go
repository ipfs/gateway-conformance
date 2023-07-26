package tests

import (
	"strconv"
	"strings"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewayBlock(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("gateway-raw-block.car")

	tests := SugarTests{
		{
			Name: "GET with format=raw param returns a raw block",
			Request: Request().
				Path("/ipfs/{{cid}}/dir", fixture.MustGetCid()).
				Query("format", "raw"),
			Response: Expect().
				Status(200).
				Body(fixture.MustGetRawData("dir")),
		},
		{
			Name: "GET with application/vnd.ipld.raw header returns a raw block",
			Request: Request().
				Path("/ipfs/{{cid}}/dir", fixture.MustGetCid()).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
				),
			Response: Expect().
				Status(200).
				Body(fixture.MustGetRawData("dir")),
		},
		{
			Name: "GET with application/vnd.ipld.raw with single range request includes correct bytes",
			Request: Request().
				Path("/ipfs/{{cid}}/dir", fixture.MustGetCid()).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
					Header("Range", "bytes=6-16"),
				),
			Response: Expect().
				Status(206).
				Body(fixture.MustGetRawData("dir")[6:17]),
		},
		// This test is disabled as per the conversation in:
		// https://github.com/ipfs/boxo/pull/369#issuecomment-1631159632
		// {
		// 	Name: "GET with application/vnd.ipld.raw with multiple range request includes correct bytes",
		// 	Request: Request().
		// 		Path("/ipfs/{{cid}}/dir/ascii.txt", fixture.MustGetCid()).
		// 		Headers(
		// 			Header("Accept", "application/vnd.ipld.raw"),
		// 			Header("Range", "bytes=6-16,0-4"),
		// 		),
		// 	Response: Expect().
		// 		Status(206).
		// 		Body(check.And(
		// 			check.Contains("Content-Range: bytes 6-16/31"),
		// 			check.Contains("Content-Type: application/vnd.ipld.raw"),
		// 			check.Contains("application"),
		// 			check.Contains("Content-Range: bytes 0-4/31"),
		// 			check.Contains("hello"),
		// 		)),
		// },
		{
			Name: "GET with application/vnd.ipld.raw header returns expected response headers",
			Request: Request().
				Path("/ipfs/{{cid}}/dir/ascii.txt", fixture.MustGetCid()).
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
				Path("/ipfs/{{cid}}/dir/ascii.txt?filename=foobar.bin", fixture.MustGetCid()).
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
				Path("/ipfs/{{cid}}/dir/ascii.txt", fixture.MustGetCid()).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
				),
			Response: Expect().
				Status(200).
				Headers(
					Header("ETag").
						Exists(),
					Header("X-IPFS-Path").
						Equals("/ipfs/{{cid}}/dir/ascii.txt", fixture.MustGetCid()),
					Header("X-IPFS-Roots").
						Contains(fixture.MustGetCid()),
					Header("Cache-Control").
						Hint("It should be public, immutable and have max-age of at least 31536000.").
						Checks(func(v string) bool {
							// TODO: port this to a regular check.
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

	RunWithSpecs(t, tests, specs.PathGatewayRaw)
}
