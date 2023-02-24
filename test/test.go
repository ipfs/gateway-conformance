package test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

func GetEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var GatewayUrl = GetEnv("GATEWAY_URL", "http://localhost:8080")

type Request struct {
	Url string
}

type Response struct {
	StatusCode int
	Body       []byte
}

type Test struct {
	Request  Request
	Response Response
}

func Run(t *testing.T, tests map[string]Test) {
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			res, err := http.Get(fmt.Sprintf("%s/%s", GatewayUrl, test.Request.Url))
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode != test.Response.StatusCode {
				t.Fatalf("Status code is not %d. It is %d", test.Response.StatusCode, res.StatusCode)
			}

			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(body, test.Response.Body) {
				t.Fatalf("Body is not '%+v'. It is: '%+v'", test.Response.Body, body)
			}
		})
	}
}
