package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/ipfs/gateway-conformance/car"
	"github.com/ipfs/gateway-conformance/test"
)

func TestGatewayBlock(t *testing.T) {
	fixture := car.MustOpenUnixfsCar("fixtures/t0117-gateway-block.car")
	tests := map[string]test.Test{
		"GET with format=raw param returns a raw block": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/dir?format=raw", fixture.MustGetCid()),
			},
			Response: test.Response{
				StatusCode: 200,
				Body:       fixture.MustGetRawData("dir"),
			},
		},
		"GET with application/vnd.ipld.raw header returns a raw block": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/dir", fixture.MustGetCid()),
				Headers: map[string]string{
					"Accept": "application/vnd.ipld.raw",
				},
			},
			Response: test.Response{
				StatusCode: 200,
				Body:       fixture.MustGetRawData("dir"),
			},
		},
		"GET with application/vnd.ipld.raw header returns expected response headers": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/dir/ascii.txt", fixture.MustGetCid()),
				Headers: map[string]string{
					"Accept": "application/vnd.ipld.raw",
				},
			},
			Response: test.Response{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"Content-Type":           "application/vnd.ipld.raw",
					"Content-Length":         fmt.Sprintf("%d", len(fixture.MustGetRawData("dir", "ascii.txt"))),
					"Content-Disposition":    regexp.MustCompile(fmt.Sprintf("attachment;\\s*filename=\"%s\\.bin", fixture.MustGetCid("dir", "ascii.txt"))),
					"X-Content-Type-Options": "nosniff",
				},
				Body: fixture.MustGetRawData("dir", "ascii.txt"),
			},
		},
		"GET with application/vnd.ipld.raw header and filename param returns expected Content-Disposition header": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/dir/ascii.txt?filename=foobar.bin", fixture.MustGetCid()),
				Headers: map[string]string{
					"Accept": "application/vnd.ipld.raw",
				},
			},
			Response: test.Response{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"Content-Disposition": regexp.MustCompile("attachment;\\s*filename=\"foobar\\.bin"),
				},
			},
		},
		"GET with application/vnd.ipld.raw header returns expected caching headers": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/dir/ascii.txt", fixture.MustGetCid()),
				Headers: map[string]string{
					"Accept": "application/vnd.ipld.raw",
				},
			},
			Response: test.Response{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"ETag":         fmt.Sprintf("\"%s.raw\"", fixture.MustGetCid("dir", "ascii.txt")),
					"X-IPFS-Path":  fmt.Sprintf("/ipfs/%s/dir/ascii.txt", fixture.MustGetCid()),
					"X-IPFS-Roots": regexp.MustCompile(fixture.MustGetCid()),
					"Cache-Control": test.Header[test.Check[string]](
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

	test.Run(t, tests)
}
