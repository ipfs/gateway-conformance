package test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"testing"
)

func GetEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var GatewayUrl = GetEnv("GATEWAY_URL", "http://127.0.0.1:8080")

type WithHintIface interface {
	GetValue() interface{}
	GetHint() string
}

type WithHint[T any] struct {
	Value T
	Hint  string
}

func (w WithHint[T]) GetValue() interface{} {
	return w.Value
}

func (w WithHint[T]) GetHint() string {
	return w.Hint
}

type Check[T any] func(T) bool

type Request struct {
	Method  string
	Url     string
	Headers map[string]string
	Body    []byte
}

type Response struct {
	StatusCode int
	Headers    map[string]interface{}
	Body       []byte
}

type Test struct {
	Request  Request
	Response Response
}

func Run(t *testing.T, tests map[string]Test) {
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			method := test.Request.Method
			if method == "" {
				method = "GET"
			}

			url := fmt.Sprintf("%s/%s", GatewayUrl, test.Request.Url)

			var body io.Reader
			if test.Request.Body != nil {
				body = bytes.NewBuffer(test.Request.Body)
			}

			// create a request
			req, err := http.NewRequest(method, url, body)
			if err != nil {
				t.Fatal(err)
			}

			// add headers
			for key, value := range test.Request.Headers {
				req.Header.Add(key, value)
			}

			// send request
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if test.Response.StatusCode != 0 {
				if res.StatusCode != test.Response.StatusCode {
					t.Fatalf("Status code is not %d. It is %d", test.Response.StatusCode, res.StatusCode)
				}
			}

			for key, value := range test.Response.Headers {
				actual := res.Header.Get(key)
				var expected string
				var hint string
				var match bool
				if w, ok := value.(WithHintIface); ok {
					value = w.GetValue()
					hint = w.GetHint()
				}
				switch v := value.(type) {
				case *regexp.Regexp:
					expected = v.String()
					match = v.MatchString(actual)
				case Check[string]:
					expected = "<Check[String]>"
					match = v(actual)
				default:
					expected = fmt.Sprintf("%v", value)
					match = actual == expected
				}
				if !match {
					t.Fatalf("Header '%s' is not '%s'. It is '%s'. %s", key, expected, actual, hint)
				}
			}

			if test.Response.Body != nil {
				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal(resBody, test.Response.Body) {
					if res.Header.Get("Content-Type") == "application/vnd.ipld.raw" {
						t.Fatalf("Body is not '%+v'. It is: '%+v'", test.Response.Body, resBody)
					} else {
						t.Fatalf("Body is not '%s'. It is: '%s'", test.Response.Body, resBody)
					}
				}
			}
		})
	}
}
