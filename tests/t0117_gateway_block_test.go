package tests

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/car"
	. "github.com/ipfs/gateway-conformance/tooling/check"
	. "github.com/ipfs/gateway-conformance/tooling/test"
)

func TestGatewayBlock(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("t0117-gateway-block.car")
	tests := []CTest{
		{
			Name: "GET with format=raw param returns a raw block",
			Request: CRequest{
				Path: fmt.Sprintf("ipfs/%s/dir?format=raw", fixture.MustGetCid()),
			},
			Response: CResponse{
				StatusCode: 200,
				Body:       fixture.MustGetRawData("dir"),
			},
		},
		{
			Name: "GET with application/vnd.ipld.raw header returns a raw block",
			Request: CRequest{
				Path: fmt.Sprintf("ipfs/%s/dir", fixture.MustGetCid()),
				Headers: map[string]string{
					"Accept": "application/vnd.ipld.raw",
				},
			},
			Response: CResponse{
				StatusCode: 200,
				Body:       fixture.MustGetRawData("dir"),
			},
		},
		{
			Name: "GET with application/vnd.ipld.raw header returns expected response headers",
			Request: CRequest{
				Path: fmt.Sprintf("ipfs/%s/dir/ascii.txt", fixture.MustGetCid()),
				Headers: map[string]string{
					"Accept": "application/vnd.ipld.raw",
				},
			},
			Response: CResponse{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"Content-Type":           "application/vnd.ipld.raw",
					"Content-Length":         IsEqual("%d", len(fixture.MustGetRawData("dir", "ascii.txt"))),
					"Content-Disposition":    Matches("attachment;\\s*filename=\"%s\\.bin", fixture.MustGetCid("dir", "ascii.txt")),
					"X-Content-Type-Options": "nosniff",
				},
				Body: fixture.MustGetRawData("dir", "ascii.txt"),
			},
		},
		{
			Name: "GET with application/vnd.ipld.raw header and filename param returns expected Content-Disposition header",
			Request: CRequest{
				Path: fmt.Sprintf("ipfs/%s/dir/ascii.txt?filename=foobar.bin", fixture.MustGetCid()),
				Headers: map[string]string{
					"Accept": "application/vnd.ipld.raw",
				},
			},
			Response: CResponse{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"Content-Disposition": Matches("attachment;\\s*filename=\"foobar\\.bin"),
				},
			},
		},
		{
			Name: "GET with application/vnd.ipld.raw header returns expected caching headers",
			Request: CRequest{
				Path: fmt.Sprintf("ipfs/%s/dir/ascii.txt", fixture.MustGetCid()),
				Headers: map[string]string{
					"Accept": "application/vnd.ipld.raw",
				},
			},
			Response: CResponse{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"ETag":         IsEqual("\"%s.raw\"", fixture.MustGetCid("dir", "ascii.txt")),
					"X-IPFS-Path":  IsEqual("/ipfs/%s/dir/ascii.txt", fixture.MustGetCid()),
					"X-IPFS-Roots": Contains(fixture.MustGetCid()),
					"Cache-Control": Checks(
						"It should be public, immutable and have max-age of at least 31536000.",
						func(v string) bool {
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
						},
					),
				},
			},
		},
	}

	Run(t, tests)
}
