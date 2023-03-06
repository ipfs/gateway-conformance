package test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/check"
)

func GetEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var GatewayUrl = strings.TrimRight(
	GetEnv("GATEWAY_URL", "http://127.0.0.1:8080"),
	"/")

type CRequest struct {
	Method               string
	RawURL               string
	DoNotFollowRedirects bool
	Url                  string
	Headers              map[string]string
	Body                 []byte
}

type CResponse struct {
	StatusCode int
	Headers    map[string]interface{}
	Body       interface{}
}

type CTest struct {
	Name     string
	Request  CRequest
	Response CResponse
}

func Run(t *testing.T, tests []CTest) {
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			method := test.Request.Method
			if method == "" {
				method = "GET"
			}

			client := &http.Client{}

			if test.Request.DoNotFollowRedirects {
				client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}
			}

			var url string
			if test.Request.RawURL != "" {
				if test.Request.Url != "" {
					t.Fatalf("Both 'RawURL' and 'Url' are set")
				}
				url = test.Request.RawURL
			} else {
				url = fmt.Sprintf("%s/%s", GatewayUrl, test.Request.Url)
			}

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
			res, err := client.Do(req)
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

				var output check.CheckOutput
				var hint string

				switch v := value.(type) {
				case check.Check[string]:
					output = v.Check(actual)
				case check.CheckWithHint[string]:
					output = v.Check.Check(actual)
					hint = v.Hint
				case string:
					output = check.IsEqual(v).Check(actual)
				default:
					t.Fatalf("Header check '%s' has an invalid type: %T", key, value)
				}

				if !output.Success {
					if hint == "" {
						t.Fatalf("Header '%s' %s", key, output.Reason)
					} else {
						t.Fatalf("Header '%s' %s (%s)", key, output.Reason, hint)
					}
				}
			}

			if test.Response.Body != nil {
				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}

				switch v := test.Response.Body.(type) {
				case check.Check[string]:
					output := v.Check(string(resBody))
					if !output.Success {
						t.Fatalf("Body %s", output.Reason)
					}
				case check.CheckWithHint[string]:
					output := v.Check.Check(string(resBody))
					if !output.Success {
						t.Fatalf("Body %s (%s)", output.Reason, v.Hint)
					}
				case string:
					if string(resBody) != v {
						t.Fatalf("Body is not '%s'. It is: '%s'", v, resBody)
					}
				case []byte:
					if !bytes.Equal(resBody, v) {

						if res.Header.Get("Content-Type") == "application/vnd.ipld.raw" {
							t.Fatalf("Body is not '%+v'. It is: '%+v'", test.Response.Body, resBody)
						} else {
							t.Fatalf("Body is not '%s'. It is: '%s'", test.Response.Body, resBody)
						}
					}
				default:
					t.Fatalf("Body check has an invalid type: %T", test.Response.Body)
				}
			}
		})
	}
}
