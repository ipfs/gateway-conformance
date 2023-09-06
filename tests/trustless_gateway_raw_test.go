package tests

import (
	"github.com/ipfs/gateway-conformance/tooling/helpers"
	"strconv"
	"strings"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling"
	"github.com/ipfs/gateway-conformance/tooling/car"
	"github.com/ipfs/gateway-conformance/tooling/specs"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestTrustlessRaw(t *testing.T) {
	tooling.LogTestGroup(t, GroupBlockCar)
	tooling.LogSpecs(t, "https://specs.ipfs.tech/http-gateways/trustless-gateway/#block-responses-application-vnd-ipld-raw")

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
	tooling.LogTestGroup(t, GroupBlockCar)
	// @lidel: "The optional entity-bytes=from:to parameter is available only for CAR requests."

	// Multi-range requests MUST conform to the HTTP semantics. The server does not
	// need to be able to support returning multiple ranges. However, it must respond
	// correctly.
	fixture := car.MustOpenUnixfsCar("gateway-raw-block.car")

	singleRangeTest := helpers.SingleRangeTestTransform(t,
		SugarTest{
			Name: "GET with application/vnd.ipld.raw with single range request includes correct bytes",
			Request: Request().
				Path("/ipfs/{{cid}}", fixture.MustGetCid("dir", "ascii.txt")).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
					Header("Range", "bytes=6-16"),
				),
			Response: Expect().
				Headers(
					Header("Content-Type").Contains("application/vnd.ipld.raw"),
				),
		},
		helpers.SimpleByteRange(6, 16, fixture.MustGetRawData("dir", "ascii.txt")[6:17]),
		fixture.MustGetRawData("dir", "ascii.txt"),
	)

	multiRangeTest := helpers.MultiRangeTestTransform(t,
		SugarTest{
			Name: "GET with application/vnd.ipld.raw with multiple range request includes correct bytes",
			Request: Request().
				Path("/ipfs/{{cid}}", fixture.MustGetCid("dir", "ascii.txt")).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
					Header("Range", "bytes=6-16,0-4"),
				),
			Response: Expect(),
		},
		helpers.ByteRanges{
			helpers.SimpleByteRange(6, 16, fixture.MustGetRawData("dir", "ascii.txt")[6:17]),
			helpers.SimpleByteRange(0, 4, fixture.MustGetRawData("dir", "ascii.txt")[0:5]),
		},
		fixture.MustGetRawData("dir", "ascii.txt"),
	)

	tests := SugarTests{
		singleRangeTest,
		multiRangeTest,
	}

	RunWithSpecs(t, tests, specs.TrustlessGatewayRaw)
}
