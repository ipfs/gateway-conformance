package test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
)

type Reporter func(t *testing.T, msg interface{}, rest ...interface{})

func runRequest(ctx context.Context, t *testing.T, test SugarTest, builder RequestBuilder) (*http.Request, *http.Response, Reporter) {
	method := builder.Method_
	if method == "" {
		method = "GET"
	}

	// Prepare a client,
	// use proxy, deal with redirects, etc.
	client := &http.Client{}
	if builder.UseProxyTunnel_ {
		if builder.Proxy_ == "" {
			t.Fatal("ProxyTunnel requires a proxy")
		}

		client = NewProxyTunnelClient(builder.Proxy_)
	} else if builder.Proxy_ != "" {
		client = NewProxyClient(builder.Proxy_)
	}

	if !builder.FollowRedirects_ {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	var res *http.Response = nil
	var req *http.Request = nil

	var localReport Reporter = func(t *testing.T, msg interface{}, rest ...interface{}) {
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
	if builder.URL_ != "" && builder.Path_ != "" {
		localReport(t, "Both 'URL' and 'Path' are set")
	}
	if builder.URL_ == "" && builder.Path_ == "" {
		localReport(t, "Neither 'URL' nor 'Path' are set")
	}
	if builder.URL_ != "" {
		url = builder.URL_
	}
	if builder.Path_ != "" {
		if builder.Path_[0] != '/' {
			localReport(t, "Path must start with '/'")
		}

		url = fmt.Sprintf("%s%s", GatewayURL, builder.Path_)
	}

	query := builder.Query_.Encode()
	if query != "" {
		url = fmt.Sprintf("%s?%s", url, query)
	}

	var body io.Reader
	if builder.Body_ != nil {
		body = bytes.NewBuffer(builder.Body_)
	}

	// create a request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatal(err)
	}

	// add headers
	for key, value := range builder.Headers_ {
		req.Header.Add(key, value)

		// https://github.com/golang/go/issues/7682
		if key == "Host" {
			req.Host = value
		}
	}

	if test.Before != nil {
		err = test.Before(req)
		if err != nil {
			localReport(t, "Before failed: %s", err)
		}
	}

	// send request
	log.Debugf("Querying %s", url)
	req = req.WithContext(ctx)
	res, err = client.Do(req)
	if err != nil {
		localReport(t, "Querying %s failed: %s", url, err)
	}

	if test.After != nil {
		err = test.After(res)
		if err != nil {
			localReport(t, "After failed: %s", err)
		}
	}

	return req, res, localReport
}
