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
	tests := map[string]test.Test{
		"GET with format=raw param returns a raw block": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/dir?format=raw", car.GetCid("fixtures/t0117-gateway-block.car")),
			},
			Response: test.Response{
				StatusCode: 200,
				Body:       car.GetRawData("fixtures/t0117-gateway-block.car", "dir"),
			},
		},
		"GET with application/vnd.ipld.raw header returns a raw block": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/dir", car.GetCid("fixtures/t0117-gateway-block.car")),
				Headers: map[string]string{
					"Accept": "application/vnd.ipld.raw",
				},
			},
			Response: test.Response{
				StatusCode: 200,
				Body:       car.GetRawData("fixtures/t0117-gateway-block.car", "dir"),
			},
		},
		"GET with application/vnd.ipld.raw header returns expected response headers": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/dir/ascii.txt", car.GetCid("fixtures/t0117-gateway-block.car")),
				Headers: map[string]string{
					"Accept": "application/vnd.ipld.raw",
				},
			},
			Response: test.Response{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"Content-Type":           "application/vnd.ipld.raw",
					"Content-Length":         fmt.Sprintf("%d", len(car.GetRawData("fixtures/t0117-gateway-block.car", "dir", "ascii.txt"))),
					"Content-Disposition":    regexp.MustCompile(fmt.Sprintf("attachment;\\s*filename=\"%s\\.bin", car.GetCid("fixtures/t0117-gateway-block.car", "dir", "ascii.txt"))),
					"X-Content-Type-Options": "nosniff",
				},
				Body: car.GetRawData("fixtures/t0117-gateway-block.car", "dir", "ascii.txt"),
			},
		},
		"GET with application/vnd.ipld.raw header and filename param returns expected Content-Disposition header": {
			Request: test.Request{
				Url: fmt.Sprintf("ipfs/%s/dir/ascii.txt?filename=foobar.bin", car.GetCid("fixtures/t0117-gateway-block.car")),
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
				Url: fmt.Sprintf("ipfs/%s/dir/ascii.txt", car.GetCid("fixtures/t0117-gateway-block.car")),
				Headers: map[string]string{
					"Accept": "application/vnd.ipld.raw",
				},
			},
			Response: test.Response{
				StatusCode: 200,
				Headers: map[string]interface{}{
					"ETag":         fmt.Sprintf("\"%s.raw\"", car.GetCid("fixtures/t0117-gateway-block.car", "dir", "ascii.txt")),
					"X-IPFS-Path":  fmt.Sprintf("/ipfs/%s/dir/ascii.txt", car.GetCid("fixtures/t0117-gateway-block.car")),
					"X-IPFS-Roots": regexp.MustCompile(car.GetCid("fixtures/t0117-gateway-block.car")),
					"Cache-Control": test.WithHint[test.Check[string]]{
						Value: func(v string) bool {
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
						Hint: "It should be public, immutable and have max-age of at least 31536000.",
					},
				},
			},
		},
	}

	test.Run(t, tests)
}
