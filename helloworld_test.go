package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func GetNode(t *testing.T, nodes []Node, p string) *Node {
	t.Helper()
	node := FindNode(nodes, p)
	if node == nil {
		t.Fatal(fmt.Errorf("node not found: %s", p))
	}
	return node
}

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
	nodes, err := ExtractCar("fixtures/dir.car")
	if err != nil {
		t.Fatal(err)
	}

	root := GetNode(t, nodes, "/")
	ascii := GetNode(t, nodes, "/dir/ascii.txt")

	// cid: `echo "helloworld" | ipfs add --inline -q`
	url := "http://localhost:8080/ipfs/" + root.Cid.String() + "/dir/ascii.txt?format=raw"
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
	if !bytes.Equal(bodyBytes, ascii.Raw) {
		t.Fatalf("Body does not contain '%+v', got: '%+v'", ascii.Raw, bodyBytes)
	}
}
