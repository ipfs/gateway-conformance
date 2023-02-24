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

type String string

func (s String) String() string {
	return string(s)
}

type StringWithHint struct {
	Value string
	Hint  string
}

func (s StringWithHint) String() string {
	return s.Value
}

type Request struct {
	Method  string
	Url     string
	Headers map[string]string
	Body    []byte
}

type Headers map[string]fmt.Stringer

type Response struct {
	StatusCode int
	Headers    Headers
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
			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			if res.StatusCode != test.Response.StatusCode {
				t.Fatalf("Status code is not %d. It is %d", test.Response.StatusCode, res.StatusCode)
			}

			for key, value := range test.Response.Headers {
				actual := res.Header.Get(key)
				expected := value.String()
				if actual != expected {
					if hint, ok := value.(StringWithHint); ok {
						t.Fatalf("Header '%s' is not '%s'. It is '%s'. Hint: %s", key, expected, actual, hint.Hint)
					} else {
						t.Fatalf("Header '%s' is not '%s'. It is '%s'", key, expected, actual)
					}
				}
			}

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(resBody, test.Response.Body) {
				t.Fatalf("Body is not '%+v'. It is: '%+v'", test.Response.Body, body)
			}
		})
	}
}
