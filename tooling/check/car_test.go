package check

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func loadFile(t *testing.T, carFilePath string) []byte {
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
	block := loadFile(t, "./_fixtures/hello_ipfs.car")

	c1 := IsCar().
		HasBlock("bafkreidfdrlkeq4m4xnxuyx6iae76fdm4wgl5d4xzsb77ixhyqwumhz244").
		HasRoot("bafkreidfdrlkeq4m4xnxuyx6iae76fdm4wgl5d4xzsb77ixhyqwumhz244")

	assert.True(t, c1.Check(block).Success)

	// invalid CID
	c2 := IsCar().
		HasBlock("bafkreiac7wncixdkhdew6wwnzya36b54t7nxcnhps377fjgtmezddnj6em")

	assert.False(t, c2.Check(block).Success)

	// missing Roots
	c3 := IsCar().
		HasBlock("bafkreidfdrlkeq4m4xnxuyx6iae76fdm4wgl5d4xzsb77ixhyqwumhz244").
		Exactly()

	assert.False(t, c3.Check(block).Success)

	// more blocks than expected
	c4 := IsCar()
	assert.True(t, c4.Check(block).Success)

	// more blocks than expected, but exact
	c5 := IsCar().
		Exactly()

	assert.False(t, c5.Check(block).Success)
}
