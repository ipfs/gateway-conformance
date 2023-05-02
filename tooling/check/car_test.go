package check

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func loadCarFile(t *testing.T, carFilePath string) []byte {
	file, err := os.Open(carFilePath)
	if err != nil {
		t.Fatalf("failed to open car file: %v", err)
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatalf("failed to read car file: %v", err)
	}

	return fileBytes
}

func TestHasFile(t *testing.T) {
	c1 := IsCar().
		HasBlock("bafkreidfdrlkeq4m4xnxuyx6iae76fdm4wgl5d4xzsb77ixhyqwumhz244").
		HasBlockWithContent("bafkreidfdrlkeq4m4xnxuyx6iae76fdm4wgl5d4xzsb77ixhyqwumhz244", []byte("Hello IPFS\n")).
		HasRoot("bafkreidfdrlkeq4m4xnxuyx6iae76fdm4wgl5d4xzsb77ixhyqwumhz244")

	// invalid CID
	c2 := IsCar().
		HasBlock("bafkreiac7wncixdkhdew6wwnzya36b54t7nxcnhps377fjgtmezddnj6em")

	// invalid content
	c3 := IsCar().
		HasBlock("bafkreidfdrlkeq4m4xnxuyx6iae76fdm4wgl5d4xzsb77ixhyqwumhz244").
		HasBlockWithContent("bafkreidfdrlkeq4m4xnxuyx6iae76fdm4wgl5d4xzsb77ixhyqwumhz244", []byte("Invalid Content\n"))

	block := loadCarFile(t, "./_fixtures/hello_ipfs.car")

	assert.Equal(t, true, c1.Check(block).Success)
	assert.Equal(t, false, c2.Check(block).Success)
	assert.Equal(t, false, c3.Check(block).Success)
}
