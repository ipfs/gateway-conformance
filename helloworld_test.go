package main

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"github.com/ipfs/gateway-conformance/car"
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
	url := "http://localhost:8080/ipfs/" + car.GetCid(t, "fixtures/dir.car", "/") + "/dir/ascii.txt?format=raw"
	res, err := http.Get(url)

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
	if !bytes.Equal(bodyBytes, car.GetRawBlock(t, "fixtures/dir.car", "/dir/ascii.txt")) {
		t.Fatalf("Body does not contain '%+v', got: '%+v'", car.GetRawBlock(t, "fixtures/dir.car", "/dir/ascii.txt"), bodyBytes)
	}
}
