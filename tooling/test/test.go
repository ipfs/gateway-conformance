package test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/check"
)

type CRequest struct {
	Method               string            `json:"method,omitempty"`
	URL                  string            `json:"url,omitempty"`
	Query                url.Values        `json:"query,omitempty"`
	Proxy                string            `json:"proxy,omitempty"`
	UseProxyTunnel       bool              `json:"useProxyTunnel,omitempty"`
	DoNotFollowRedirects bool              `json:"doNotFollowRedirects,omitempty"`
	Path                 string            `json:"path,omitempty"`
	Subdomain            string            `json:"subdomain,omitempty"`
	Headers              map[string]string `json:"headers,omitempty"`
	Body                 []byte            `json:"body,omitempty"`
}

type CResponse struct {
	StatusCode int                    `json:"statusCode,omitempty"`
	Headers    map[string]interface{} `json:"headers,omitempty"`
	Body       interface{}            `json:"body,omitempty"`
}

type SugarTest struct {
	Name     string
	Hint     string
	Request  RequestBuilder
	Response ExpectBuilder
}

type SugarTests []SugarTest

func Run(t *testing.T, tests SugarTests) {
	// NewDialer()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			request := test.Request.Request()
			response := test.Response.Response()

			method := request.Method
			if method == "" {
				method = "GET"
			}

			// Prepare a client,
			// use proxy, deal with redirects, etc.
			client := &http.Client{}
			if request.UseProxyTunnel {
				if request.Proxy == "" {
					t.Fatal("ProxyTunnel requires a proxy")
				}

				client = NewProxyTunnelClient(request.Proxy)
			} else if request.Proxy != "" {
				client = NewProxyClient(request.Proxy)
			}

			if request.DoNotFollowRedirects {
				client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				}
			}

			var res *http.Response = nil
			var req *http.Request = nil

			localReport := func(t *testing.T, msg interface{}, rest ...interface{}) {
				var err error
				switch msg := msg.(type) {
				case string:
					err = fmt.Errorf(msg, rest...)
				case error:
					err = msg
				default:
					panic("msg must be string or error")
				}

				report(t, test, req, res, err)
			}

			var url string
			if request.URL != "" && request.Path != "" {
				localReport(t, "Both 'URL' and 'Path' are set")
			}
			if request.URL == "" && request.Path == "" {
				localReport(t, "Neither 'URL' nor 'Path' are set")
			}
			if request.URL != "" {
				url = request.URL
			}
			if request.Path != "" {
				url = fmt.Sprintf("%s/%s", GatewayURL, request.Path)
			}

			query := request.Query.Encode()
			if query != "" {
				url = fmt.Sprintf("%s?%s", url, query)
			}

			var body io.Reader
			if request.Body != nil {
				body = bytes.NewBuffer(request.Body)
			}

			// create a request
			req, err := http.NewRequest(method, url, body)
			if err != nil {
				t.Fatal(err)
			}

			// add headers
			for key, value := range request.Headers {
				req.Header.Add(key, value)

				// https://github.com/golang/go/issues/7682
				if key == "Host" {
					req.Host = value
				}
			}

			// send request
			log.Debugf("Querying %s", url)
			res, err = client.Do(req)
			if err != nil {
				localReport(t, "Querying %s failed: %s", url, err)
			}

			if response.StatusCode != 0 {
				if res.StatusCode != response.StatusCode {
					localReport(t, "Status code is not %d. It is %d", response.StatusCode, res.StatusCode)
				}
			}

			for key, value := range response.Headers {
				t.Run(fmt.Sprintf("Header %s", key), func(t *testing.T) {
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
						localReport(t, "Header check '%s' has an invalid type: %T", key, value)
					}

					if !output.Success {
						if hint == "" {
							localReport(t, "Header '%s' %s", key, output.Reason)
						} else {
							localReport(t, "Header '%s' %s (%s)", key, output.Reason, hint)
						}
					}
				})
			}

			if response.Body != nil {
				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				if err != nil {
					localReport(t, err)
				}

				switch v := response.Body.(type) {
				case check.Check[string]:
					output := v.Check(string(resBody))
					if !output.Success {
						localReport(t, "Body %s", output.Reason)
					}
				case check.CheckWithHint[string]:
					output := v.Check.Check(string(resBody))
					if !output.Success {
						localReport(t, "Body %s (%s)", output.Reason, v.Hint)
					}
				case string:
					if string(resBody) != v {
						localReport(t, "Body is not '%s'. It is: '%s'", v, resBody)
					}
				case []byte:
					if !bytes.Equal(resBody, v) {
						if res.Header.Get("Content-Type") == "application/vnd.ipld.raw" {
							localReport(t, "Body is not '%+v'. It is: '%+v'", response.Body, resBody)
						} else {
							localReport(t, "Body is not '%s'. It is: '%s'", response.Body, resBody)
						}
					}
				default:
					localReport(t, "Body check has an invalid type: %T", response.Body)
				}
			}
		})
	}
}
