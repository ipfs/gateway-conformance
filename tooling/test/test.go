package test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling/check"
)

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
			method := test.Request.Method_
			if method == "" {
				method = "GET"
			}

			// Prepare a client,
			// use proxy, deal with redirects, etc.
			client := &http.Client{}
			if test.Request.UseProxyTunnel_ {
				if test.Request.Proxy_ == "" {
					t.Fatal("ProxyTunnel requires a proxy")
				}

				client = NewProxyTunnelClient(test.Request.Proxy_)
			} else if test.Request.Proxy_ != "" {
				client = NewProxyClient(test.Request.Proxy_)
			}

			if test.Request.DoNotFollowRedirects_ {
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
			if test.Request.URL_ != "" && test.Request.Path_ != "" {
				localReport(t, "Both 'URL' and 'Path' are set")
			}
			if test.Request.URL_ == "" && test.Request.Path_ == "" {
				localReport(t, "Neither 'URL' nor 'Path' are set")
			}
			if test.Request.URL_ != "" {
				url = test.Request.URL_
			}
			if test.Request.Path_ != "" {
				url = fmt.Sprintf("%s/%s", GatewayURL, test.Request.Path_)
			}

			query := test.Request.Query_.Encode()
			if query != "" {
				url = fmt.Sprintf("%s?%s", url, query)
			}

			var body io.Reader
			if test.Request.Body_ != nil {
				body = bytes.NewBuffer(test.Request.Body_)
			}

			// create a request
			req, err := http.NewRequest(method, url, body)
			if err != nil {
				t.Fatal(err)
			}

			// add headers
			for key, value := range test.Request.Headers_ {
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

			if test.Response.StatusCode_ != 0 {
				if res.StatusCode != test.Response.StatusCode_ {
					localReport(t, "Status code is not %d. It is %d", test.Response.StatusCode_, res.StatusCode)
				}
			}

			for _, header := range test.Response.Headers_ {
				t.Run(fmt.Sprintf("Header %s", header.Key_), func(t *testing.T) {
					actual := res.Header.Get(header.Key_)
					output := header.Check_.Check(actual)

					if !output.Success {
						if header.Hint_ == "" {
							localReport(t, "Header '%s' %s", header.Key_, output.Reason)
						} else {
							localReport(t, "Header '%s' %s (%s)", header.Key_, output.Reason, header.Hint_)
						}
					}
				})
			}

			if test.Response.Body_ != nil {
				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				if err != nil {
					localReport(t, err)
				}

				var output check.CheckOutput

				switch v := test.Response.Body_.(type) {
				case check.Check[string]:
					output = v.Check(string(resBody))
				case check.Check[[]byte]:
					output = v.Check(resBody)
				case string:
					output = check.IsEqual(v).Check(string(resBody))
				case []byte:
					output = check.IsEqualBytes(v).Check(resBody)
				default:
					output = check.CheckOutput{
						Success: false,
						Reason:  fmt.Sprintf("Body check has an invalid type: %T", test.Response.Body_),
					}
				}

				if !output.Success {
					if output.Hint == "" {
						localReport(t, "Body %s", output.Reason)
					} else {
						localReport(t, "Body %s (%s)", output.Reason, output.Hint)
					}
				}
			}
		})
	}
}
