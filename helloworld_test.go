package main

import (
	"io"
	"net/http"
	"testing"
)

/**
 * "Test HTTP Gateway Raw Block (application/vnd.ipld.raw) Support": {
 *  tests: {
 *    "GET with format=raw param returns a raw block": {
 *      url: `/ipfs/${Fixture.get("dir").getRootCID()}/dir?format=raw`,
 *      expect: [200, Fixture.get("dir").getString("dir")],
 *    },
**/

// # Test HTTP Gateway Raw Block (application/vnd.ipld.raw) Support
// ## GET with format=raw param returns a raw block
func TestGETWithFormatRawParamReturnsARawBlock(t *testing.T) {
	// cid: `echo "helloworld" | ipfs add --inline -q`
	res, err := http.Get("http://localhost:8080/ipfs/Qmckhu9X5A4K6wNzQGSrDRoHPhSpbmUELTFRdYjeQZx1M3/dir/ascii.txt?format=txt")

	if err != nil {
		t.Fatal(err)
	}

	if res.StatusCode != 200 {
		t.Fatalf("Status code is not 200. It is %d", res.StatusCode)
	}

	// check that the body contains "helloworld"
	defer res.Body.Close()
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	bodyString := string(bodyBytes)
	expected := "goodbye application/vnd.ipld.raw\n"
	if bodyString != expected {
		t.Fatalf("Body does not contain '%+v', got: '%+v'", expected, bodyString)
	}
}