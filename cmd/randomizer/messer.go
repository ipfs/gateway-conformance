package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

func ResponseMesser(resp *http.Response) error {
	// Swap two random bytes in the response headers
	for k, v := range resp.Header {
		// ignore most headers
		kk := strings.ToLower(k)
		if _, ok := messedHeaders[kk]; !ok {
			fmt.Println("could not find:", kk, "in messedHeaders")
			continue
		}

		swapRandomStrings(v)
		fmt.Println("messed:", k, v)
		resp.Header[k] = v
	}

	// randomly add headers that do not exists
	for k, v := range addedHeaders {
		if rand.Intn(10) > 1 || resp.Header.Get(k) != "" {
			continue
		}

		resp.Header[k] = []string{v}
	}

	// Swap two random bytes in the response body
	length := -1
	var err error

	// if resp has content length header,
	// extract the new length and store it.
	if resp.Header.Get("Content-Length") != "" {
		cl := resp.Header.Get("Content-Length")
		length, err = strconv.Atoi(cl)
		if err != nil {
			panic(err)
		}
	}

	swapRandomBytesReader := &swapRandomBytesReader{
		Reader: resp.Body,
	}

	resp.Body = MyLimitReader(swapRandomBytesReader, int64(length))

	resp.StatusCode = resp.StatusCode + 1

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

func messRandomNumbers(s []string) {
	for i := range s {
		s[i] = messRandomNumber(s[i])
	}
}

func messRandomNumber(s string) string {
	d, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}
	r := rand.Int63n(d)

	return strconv.FormatInt(r, 10)
}
