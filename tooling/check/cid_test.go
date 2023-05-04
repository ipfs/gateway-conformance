package check

import (
	"testing"

	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/assert"
)

func makeCID(s string) cid.Cid {
	hash, err := mh.Sum([]byte(s), mh.SHA2_256, -1)
	if err != nil {
		panic(err)
	}

	// Create a CID using the multihash
	cidV1 := cid.NewCidV1(cid.Raw, hash)

	return cidV1
}

func cids(s ...string) []cid.Cid {
	out := make([]cid.Cid, len(s))
	for i, str := range s {
		out[i] = makeCID(str)
	}
	return out
}

func TestCIDContains(t *testing.T) {
	a := cids("hello")
	b := cids()
	assert.True(t, CidSetContains(a, b).Success)

	a = cids("hello")
	b = cids("hello")
	assert.True(t, CidSetContains(a, b).Success)

	a = cids("hello")
	b = cids("world")
	assert.False(t, CidSetContains(a, b).Success)

	a = cids("hello", "world")
	b = cids("hello")
	assert.True(t, CidSetContains(a, b).Success)
}

func TestCIDEquals(t *testing.T) {
	a := cids("hello")
	b := cids()
	assert.False(t, CidSetEquals(a, b).Success)

	a = cids("hello")
	b = cids("hello")
	assert.True(t, CidSetEquals(a, b).Success)

	a = cids("hello", "world")
	b = cids("world", "hello")
	assert.True(t, CidSetEquals(a, b).Success)

	a = cids("hello")
	b = cids("world")
	assert.False(t, CidSetEquals(a, b).Success)

	a = cids("hello", "world")
	b = cids("hello")
	assert.False(t, CidSetEquals(a, b).Success)
}

func TestCidArrayEquals(t *testing.T) {
	a := cids("hello")
	b := cids()
	assert.False(t, CidArrayEquals(a, b).Success)

	a = cids("hello")
	b = cids("hello")
	assert.True(t, CidArrayEquals(a, b).Success)

	a = cids("hello", "world")
	b = cids("world", "hello")
	assert.False(t, CidArrayEquals(a, b).Success)

	a = cids("hello", "world")
	b = cids("hello", "world", "foo")
	assert.False(t, CidArrayEquals(a, b).Success)

	a = cids("hello", "world")
	b = cids("hello", "world")
	assert.True(t, CidArrayEquals(a, b).Success)

	a = cids("hello")
	b = cids("world")
	assert.False(t, CidArrayEquals(a, b).Success)

	a = cids("hello", "world")
	b = cids("hello")
	assert.False(t, CidArrayEquals(a, b).Success)
}

func TestCidArrayContains(t *testing.T) {
	a := cids("hello")
	b := cids()
	assert.True(t, CidOrderedSubsetContains(a, b).Success)

	a = cids("hello")
	b = cids("hello")
	assert.True(t, CidOrderedSubsetContains(a, b).Success)

	a = cids("hello", "world")
	b = cids("world", "hello")
	assert.False(t, CidOrderedSubsetContains(a, b).Success)

	a = cids("hello", "world")
	b = cids("hello", "world", "foo")
	assert.False(t, CidOrderedSubsetContains(a, b).Success)

	a = cids("hello", "world")
	b = cids("hello", "world")
	assert.True(t, CidOrderedSubsetContains(a, b).Success)

	a = cids("hello")
	b = cids("world")
	assert.False(t, CidOrderedSubsetContains(a, b).Success)

	a = cids("hello", "world")
	b = cids("hello")
	assert.True(t, CidOrderedSubsetContains(a, b).Success)

	a = cids("this", "hello", "world", "is", "a", "test")
	b = cids("hello", "a", "test")
	assert.True(t, CidOrderedSubsetContains(a, b).Success)
}
