package tests

import (
	"strconv"
	"strings"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewayBlock(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0117-gateway-block.car")

	tests := SugarTests{
		{
			Name: "GET with format=raw param returns a raw block",
			Request: Request().
				Path("ipfs/%s/dir", fixture.MustGetCid()).
				Query("format", "raw"),
			Response: Expect().
				Status(200).
				Body(fixture.MustGetRawData("dir")),
		},
		{
			Name: "GET with application/vnd.ipld.raw header returns a raw block",
			Request: Request().
				Path("ipfs/%s/dir", fixture.MustGetCid()).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
				),
			Response: Expect().
				Status(200).
				Body(fixture.MustGetRawData("dir")),
		},
		{
			// This was the original shell test:
			// test_expect_success "GET response for application/vnd.ipld.raw has expected Content-Type" '
			// curl -svX GET -H "Accept: application/vnd.ipld.raw" "http://127.0.0.1:$GWAY_PORT/ipfs/$ROOT_DIR_CID/dir/ascii.txt" >/dev/null 2>curl_output &&
			// test_should_contain "< Content-Type: application/vnd.ipld.raw" curl_output
			// '
			// test_expect_success "GET response for application/vnd.ipld.raw includes Content-Length" '
			// BYTES=$(ipfs block get $FILE_CID | wc --bytes)
			// test_should_contain "< Content-Length: $BYTES" curl_output
			// '
			// test_expect_success "GET response for application/vnd.ipld.raw includes Content-Disposition" '
			// test_should_contain "< Content-Disposition: attachment\; filename=\"${FILE_CID}.bin\"" curl_output
			// '
			// test_expect_success "GET response for application/vnd.ipld.raw includes nosniff hint" '
			// test_should_contain "< X-Content-Type-Options: nosniff" curl_output
			// '
			//
			// TODO: complete this requests + response
			// bonus point for testing the body as well.
			Name: "GET with application/vnd.ipld.raw header returns expected response & headers",
			Request: Request().
				Path("ipfs/%s/dir/ascii.txt", fixture.MustGetCid()),
			Response: Expect().
				Status(200),
		},
		{
			Name: "GET with application/vnd.ipld.raw header and filename param returns expected Content-Disposition header with custom filename",
			Request: Request().
				Path("ipfs/%s/dir/ascii.txt?filename=foobar.bin", fixture.MustGetCid()).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
				),
			Response: Expect().
				Status(200).
				Headers(
					Header("Content-Disposition").
						Matches("attachment;\\s*filename=\"foobar\\.bin"),
				),
		},
		{
			Name: "GET with application/vnd.ipld.raw header returns expected caching headers",
			Request: Request().
				Path("ipfs/%s/dir/ascii.txt", fixture.MustGetCid()).
				Headers(
					Header("Accept", "application/vnd.ipld.raw"),
				),
			Response: Expect().
				Status(200).
				Headers(
					Header("ETag").
						Equals("\"%s.raw\"", fixture.MustGetCid("dir", "ascii.txt")),
					Header("X-IPFS-Path").
						Equals("/ipfs/%s/dir/ascii.txt", fixture.MustGetCid()),
					Header("X-IPFS-Roots").
						Contains(fixture.MustGetCid()),
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

	Run(t, tests)
}
