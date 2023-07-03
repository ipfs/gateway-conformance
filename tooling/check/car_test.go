package check

import (
	"io"
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

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("failed to read car file: %v", err)
	}

	return fileBytes
}

func TestHasFile(t *testing.T) {
	block := loadFile(t, "./_fixtures/dag.car")

	// › npx ipfs-car ls ./dag.car --verbose
	// bafybeidlbwbu73tbjr3atntjz4lq5ego5w2uyof35vvwcnheaftzi3rndu     -       .
	// bafkreidw23elffhagxz3oi6ctoibqouzfowfn3bwcvq2yzgd5n5h4gjyou     726     ./a-file.txt
	// bafkreiaeqsxxqwmsnhzhrlyr2udn25hpj24bs7gzcgkhbrkmhcuikcgh4a     29998   ./b-file.txt
	// bafybeiaq6e55xratife7s5cmzjcmwy4adzzlk74sbdpfcq72gus6cweeeq     -       ./subdir
	// bafkreihdhgb5vyuqu7jssreyo3h567obewtqq37fi5hr2w4um5icacry7m     30999   ./subdir/leaf.txt
	c := IsCar().
		HasBlock("bafybeidlbwbu73tbjr3atntjz4lq5ego5w2uyof35vvwcnheaftzi3rndu").
		HasRoot("bafybeidlbwbu73tbjr3atntjz4lq5ego5w2uyof35vvwcnheaftzi3rndu")
	assert.True(t, c.Check(block).Success)

	c = IsCar().
		HasBlocks("bafybeidlbwbu73tbjr3atntjz4lq5ego5w2uyof35vvwcnheaftzi3rndu",
			"bafkreidw23elffhagxz3oi6ctoibqouzfowfn3bwcvq2yzgd5n5h4gjyou",
			"bafkreiaeqsxxqwmsnhzhrlyr2udn25hpj24bs7gzcgkhbrkmhcuikcgh4a",
			"bafybeiaq6e55xratife7s5cmzjcmwy4adzzlk74sbdpfcq72gus6cweeeq",
			"bafkreihdhgb5vyuqu7jssreyo3h567obewtqq37fi5hr2w4um5icacry7m",
		).
		HasRoots(
			"bafybeidlbwbu73tbjr3atntjz4lq5ego5w2uyof35vvwcnheaftzi3rndu",
		).
		Exactly()
	assert.True(t, c.Check(block).Success)

	c = IsCar().
		HasBlocks(
			// Note the order here
			"bafkreidw23elffhagxz3oi6ctoibqouzfowfn3bwcvq2yzgd5n5h4gjyou",
			"bafkreiaeqsxxqwmsnhzhrlyr2udn25hpj24bs7gzcgkhbrkmhcuikcgh4a",
			"bafkreihdhgb5vyuqu7jssreyo3h567obewtqq37fi5hr2w4um5icacry7m",
			"bafybeiaq6e55xratife7s5cmzjcmwy4adzzlk74sbdpfcq72gus6cweeeq",
			"bafybeidlbwbu73tbjr3atntjz4lq5ego5w2uyof35vvwcnheaftzi3rndu",
		).
		HasRoots(
			"bafybeidlbwbu73tbjr3atntjz4lq5ego5w2uyof35vvwcnheaftzi3rndu",
		).
		InThatOrder().
		Exactly()
	assert.True(t, c.Check(block).Success)

	c = IsCar().
		HasBlocks(
			"bafkreiaeqsxxqwmsnhzhrlyr2udn25hpj24bs7gzcgkhbrkmhcuikcgh4a",
			"bafkreihdhgb5vyuqu7jssreyo3h567obewtqq37fi5hr2w4um5icacry7m",
			"bafybeiaq6e55xratife7s5cmzjcmwy4adzzlk74sbdpfcq72gus6cweeeq",
			"bafybeidlbwbu73tbjr3atntjz4lq5ego5w2uyof35vvwcnheaftzi3rndu",
			"bafkreidw23elffhagxz3oi6ctoibqouzfowfn3bwcvq2yzgd5n5h4gjyou",
		).
		HasRoots(
			"bafybeidlbwbu73tbjr3atntjz4lq5ego5w2uyof35vvwcnheaftzi3rndu",
		).
		InThatOrder().
		Exactly()
	assert.False(t, c.Check(block).Success)

	c = IsCar().
		HasBlock("bafybeidlbwbu73tbjr3atntjz4lq5ego5w2uyof35vvwcnheaftzi3rndu").
		HasRoot("bafkreidw23elffhagxz3oi6ctoibqouzfowfn3bwcvq2yzgd5n5h4gjyou").
		Exactly()
	assert.False(t, c.Check(block).Success)

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
