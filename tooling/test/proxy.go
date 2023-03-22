package test

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
)

func NewProxyTunnelClient(proxyURL string) *http.Client {
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		panic(err)
	}

	dialer := &net.Dialer{}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxy),
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Connect to the proxy server
			conn, err := dialer.DialContext(ctx, "tcp", proxy.Host)
			if err != nil {
				return nil, err
			}

			// Send the CONNECT request to establish a tunnel
			connectReq := &http.Request{
				Method: "CONNECT",
				URL:    &url.URL{Opaque: addr},
				Host:   addr,
				Header: make(http.Header),
			}
			if err := connectReq.Write(conn); err != nil {
				conn.Close()
				return nil, err
			}

			// Read the CONNECT response from the proxy
			resp, err := http.ReadResponse(bufio.NewReader(conn), connectReq)
			if err != nil {
				conn.Close()
				return nil, err
			}
			if resp.StatusCode != 200 {
				conn.Close()
				return nil, fmt.Errorf("proxy error: %v", resp.Status)
			}

			return conn, nil
		},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: transport,
	}

	return client
}

func NewProxyClient(proxyURL string) *http.Client {
	proxy, err := url.Parse(proxyURL)
	if err != nil {
		panic(err)
	}

	transport := &http.Transport{
		Proxy:             http.ProxyURL(proxy),
		ForceAttemptHTTP2: false,
	}

	client := &http.Client{
		Transport: transport,
	}

	return client
}
