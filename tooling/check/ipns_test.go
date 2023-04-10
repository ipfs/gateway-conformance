package check

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIpnsCanOpenARecord(t *testing.T) {
	path := "../../fixtures/t0124/k51qzi5uqu5dhjjqhpcuvdnskq4mz84a2xg1rpqzi6s5460q2612egkfjsk42x.ipns-record"

	// read file:
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	check := IsIPNSKey().IsValid()
	// ipfs name inspect --verify $IPNS_KEY < curl_output_filename > verify_output &&
	output := check.Check(data)

	assert.True(t, output.Success)
}
