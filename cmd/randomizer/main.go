package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// list of all headers we want to mess with.
// It's a map from string to bool to make finding a header quick
type Messer func([]string)

var messedHeaders = map[string]Messer{
	"Content-Type":            swapRandomStrings,
	"Content-Length":          messRandomNumbers,
	"Content-Encoding":        swapRandomStrings,
	"Content-Language":        swapRandomStrings,
	"Content-Location":        swapRandomStrings,
	"Content-MD5":             swapRandomStrings,
	"Content-Range":           swapRandomStrings,
	"Content-Disposition":     swapRandomStrings,
	"Content-Features":        swapRandomStrings,
	"Content-Security-Policy": swapRandomStrings,
	"Cache-Control":           swapRandomStrings,
	"X-Ipfs-Path":             swapRandomStrings,
	"X-Ipfs-Roots":            swapRandomStrings,
	"X-Content-Type-Options":  swapRandomStrings,
	"Etag":                    swapRandomStrings,
	"Location":                swapRandomStrings,
	"Accept-Ranges":           swapRandomStrings,
	"If-None-Match":           swapRandomStrings,
}

var addedHeaders = map[string]string{
	"Cache-Control": "no-cache",
}

func init() {
	// go through all the headers and just switch to lowercase keys:
	for k, v := range messedHeaders {
		kk := strings.ToLower(k)
		fmt.Println("Replacing:", k, "with:", kk)
		delete(messedHeaders, k)
		messedHeaders[kk] = v
	}

	// print all kv in messedHeaders
	for k, v := range messedHeaders {
		fmt.Println("messedHeaders:", k, v)
	}
}

func main() {
	// Parse command-line arguments for target URL and proxy address
	var targetUrlStr, proxyAddr string
	flag.StringVar(&targetUrlStr, "target", "", "Target URL to proxy requests to")
	flag.StringVar(&proxyAddr, "proxy", "", "Address to listen for proxy requests")
	flag.Parse()

	if targetUrlStr == "" || proxyAddr == "" {
		fmt.Println("Usage: randomizer -target <target-url> -proxy <proxy-addr>")
		return
	}

	// Parse the target URL and create a reverse proxy
	targetUrl, err := url.Parse(targetUrlStr)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(targetUrl)

	// Modify the response headers and body
	proxy.ModifyResponse = ResponseMesser

	// Start the reverse proxy server on the given address
	fmt.Printf("Listening for proxy requests on %s\n", proxyAddr)
	err = http.ListenAndServe(proxyAddr, proxy)
	if err != nil {
		panic(err)
	}
}
