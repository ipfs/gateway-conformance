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
			request := test.Request
			response := test.Response

			method := request.Method_
			if method == "" {
				method = "GET"
			}

			// Prepare a client,
			// use proxy, deal with redirects, etc.
			client := &http.Client{}
			if request.UseProxyTunnel_ {
				if request.Proxy_ == "" {
					t.Fatal("ProxyTunnel requires a proxy")
				}

				client = NewProxyTunnelClient(request.Proxy_)
			} else if request.Proxy_ != "" {
				client = NewProxyClient(request.Proxy_)
			}

			if request.DoNotFollowRedirects_ {
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
			if request.URL_ != "" && request.Path_ != "" {
				localReport(t, "Both 'URL' and 'Path' are set")
			}
			if request.URL_ == "" && request.Path_ == "" {
				localReport(t, "Neither 'URL' nor 'Path' are set")
			}
			if request.URL_ != "" {
				url = request.URL_
			}
			if request.Path_ != "" {
				url = fmt.Sprintf("%s/%s", GatewayURL, request.Path_)
			}

			query := request.Query_.Encode()
			if query != "" {
				url = fmt.Sprintf("%s?%s", url, query)
			}

			var body io.Reader
			if request.Body_ != nil {
				body = bytes.NewBuffer(request.Body_)
			}

			// create a request
			req, err := http.NewRequest(method, url, body)
			if err != nil {
				t.Fatal(err)
			}

			// add headers
			for key, value := range request.Headers_ {
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

			if response.StatusCode_ != 0 {
				if res.StatusCode != response.StatusCode_ {
					localReport(t, "Status code is not %d. It is %d", response.StatusCode_, res.StatusCode)
				}
			}

			for _, header := range response.Headers_ {
				t.Run(fmt.Sprintf("Header %s", header.Key_), func(t *testing.T) {
					actual := res.Header.Get(header.Key_)
					output := header.Check_.Check(actual)
					hint := header.Hint_

					if !output.Success {
						if hint == "" {
							localReport(t, "Header '%s' %s", header.Key_, output.Reason)
						} else {
							localReport(t, "Header '%s' %s (%s)", header.Key_, output.Reason, hint)
						}
					}
				})
			}

			if response.Body_ != nil {
				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				if err != nil {
					localReport(t, err)
				}

				switch v := response.Body_.(type) {
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
							localReport(t, "Body is not '%+v'. It is: '%+v'", response.Body_, resBody)
						} else {
							localReport(t, "Body is not '%s'. It is: '%s'", response.Body_, resBody)
						}
					}
				default:
					localReport(t, "Body check has an invalid type: %T", response.Body_)
				}
			}
		})
	}
}
