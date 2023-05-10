package ipns

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIpnsCanOpenARecord(t *testing.T) {
	path := "./_fixtures/k51qzi5uqu5dgh7y9l90nqs6tvnzcm9erbt8fhzg3fu79p5qt9zb2izvfu51ki.ipns-record"

	// read file:
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	check := IsIPNSKey("k51qzi5uqu5dgh7y9l90nqs6tvnzcm9erbt8fhzg3fu79p5qt9zb2izvfu51ki").IsValid()
	// ipfs name inspect --verify $IPNS_KEY < curl_output_filename > verify_output &&
	output := check.Check(data)

	assert.True(t, output.Success)
}
