package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

// list of all headers we want to mess with.
// It's a map from string to bool to make finding a header quick
type Messer func([]string)

var messedHeaders = map[string]Messer {
	"Content-Type": swapRandomStrings,
	// "Content-Length": swapRandomNumbers,
	"Content-Encoding": swapRandomStrings,
	"Content-Language": swapRandomStrings,
	"Content-Location": swapRandomStrings,
	"Content-MD5": swapRandomStrings,
	"Content-Range": swapRandomStrings,
	"Content-Disposition": swapRandomStrings,
	"Content-Features": swapRandomStrings,
	"Content-Security-Policy": swapRandomStrings,
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
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Swap two random bytes in the response headers
		for k, v := range resp.Header {
			// ignore most headers
			if _, ok := messedHeaders[k]; !ok {
				continue
			}

			swapRandomStrings(v)
			fmt.Println("messed:", k, v)
			resp.Header[k] = v
		}
		// Swap two random bytes in the response body
		swapRandomBytesReader := &swapRandomBytesReader{Reader: resp.Body}
		resp.Body = swapRandomBytesReader

		return nil
	}

	// Start the reverse proxy server on the given address
	fmt.Printf("Listening for proxy requests on %s\n", proxyAddr)
	err = http.ListenAndServe(proxyAddr, proxy)
	if err != nil {
		panic(err)
	}
}

// A reader that swaps two random bytes in the input
type swapRandomBytesReader struct {
	io.Reader
}

func (r *swapRandomBytesReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)

	if err != nil && err != io.EOF {
		return n, err
	}
	
	fmt.Println("before:", string(p[:n]))
	shuffleBytes(p[:n])
	fmt.Println("after:", string(p[:n]))
	// swapRandomBytes(p[:n])
	return n, err
}

func (r *swapRandomBytesReader) Close() error {
	if closer, ok := r.Reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func randPlaces(max int) (int, int) {
	i := rand.Intn(max)
	j := rand.Intn(max)
	if i == j {
		j = (j + 1) % max
	}

	if i < j {
		return i, j
	}
	return j, i
}

// Shuffle bytes
func shuffleBytes(b []byte) {
	for i := range b {
		j := rand.Intn(i + 1)
		b[i], b[j] = b[j], b[i]
	}
}

// Swaps two random bytes in the input
func swapRandomBytes(b []byte) {
	if len(b) < 2 {
		return
	}

	i, j := randPlaces(len(b))
	b[i], b[j] = b[j], b[i]
}

// Swaps two random characters in the input string
func swapRandomStrings(s []string) {
	for i := range s {
		s[i] = swapRandomString(s[i])
	}
}

func swapRandomString(s string) string {
	if len(s) < 2 {
		return s
	}

	i, j := randPlaces(len(s))

	return s[:i] + string(s[j]) + s[i+1:j] + string(s[i]) + s[j+1:]
}

func swapRandomNumbers(s []string) {
	for i := range s {
		s[i] = swapRandomNumber(s[i])
	}
}

func swapRandomNumber(s string) string {
	d, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	r := rand.Int63n(d)

	return strconv.FormatInt(r, 10)
}

