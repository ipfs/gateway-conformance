package test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ipfs/gateway-conformance/tooling"
)

type Reporter func(t *testing.T, msg interface{}, rest ...interface{})

func runRequest(ctx context.Context, t *testing.T, test SugarTest, builder RequestBuilder) (*http.Request, *http.Response, Reporter) {
	method := builder.Method_
	if method == "" {
		method = "GET"
	}

	// Prepare a client,
	client := &http.Client{}

	// HTTP proxy tests require additional prep
	if builder.UseProxyTunnel_ {
		if builder.Proxy_ == "" {
			t.Fatal("ProxyTunnel requires a proxy")
		}

		client = NewProxyTunnelClient(builder.Proxy_)
	} else if builder.Proxy_ != "" {
		client = NewProxyClient(builder.Proxy_)
	}

	// Handle redirect tests
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
	if builder.Path_ == "" {
		localReport(t, "'Path' is not set")
	}
	if builder.Path_ != "" {
		if builder.Proxy_ != "" && !builder.UseProxyTunnel_ {
			// plain HTTP proxy test uses custom client, and the Path is the full URL
			// to be used in the request
			if !strings.HasPrefix(builder.Path_, "http") {
				t.Fatalf("plain Proxy tests require requested Path to be full URL starting with http. builder.Path_ was %q", builder.Path_)
			}
			// plain proxy requests use Path as-is
			url = builder.Path_
		} else {
			// no HTTP proxy, make a regular HTTP request for Path at GatewayURL (+ optional Host header)
			if builder.Path_[0] != '/' {
				localReport(t, "When proxy mode is not used, the Path must start with '/'")
			}
			// regular requests attach Path to gateway endpoint URL
			url = fmt.Sprintf("%s%s", strings.TrimRight(GatewayURL().String(), "/"), builder.Path_)
		}
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

	// Set meaningful User-Agent, if custom one was not set by a test
	if _, exists := builder.Headers_["User-Agent"]; !exists {
		req.Header.Set("User-Agent", "ipfs/gateway-conformance/"+tooling.Version)
	}

	// Send request
	log.Debugf("Querying %s", url)
	req = req.WithContext(ctx)

	res, err = client.Do(req)
	if err != nil {
		localReport(t, "Querying %s failed: %s", url, err)
	}

	return req, res, localReport
}
