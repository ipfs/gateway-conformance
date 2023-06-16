package tests

import (
	"strconv"
	"strings"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestTrustlessRaw(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0117-gateway-block.car")

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
						Equals("{{ length }}", len(fixture.MustGetRawData("dir", "ascii.txt"))),
					Header("Content-Disposition").
						Matches(`attachment;\s*filename=".*\.bin"`),
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
						Matches(`attachment;\s*filename="foobar\.bin`),
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
						Hint("Etag must be present for caching purposes").
						Not().IsEmpty(),
					Header("X-IPFS-Path").
						Equals("/ipfs/{{cid}}", fixture.MustGetCid("dir", "ascii.txt")),
					Header("X-IPFS-Roots").
						Contains(fixture.MustGetCid("dir", "ascii.txt")),
					Header("Cache-Control").
						Hint("It should be public, immutable and have max-age of at least 31536000.").
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
