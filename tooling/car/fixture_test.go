package car

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNodes(t *testing.T) {
	f := MustOpenUnixfsCar("./_fixtures/dag.car")

	// › npx ipfs-car ls ./dag.car --verbose
	// bafybeidlbwbu73tbjr3atntjz4lq5ego5w2uyof35vvwcnheaftzi3rndu     -       .
	// bafkreidw23elffhagxz3oi6ctoibqouzfowfn3bwcvq2yzgd5n5h4gjyou     726     ./a-file.txt
	// bafkreiaeqsxxqwmsnhzhrlyr2udn25hpj24bs7gzcgkhbrkmhcuikcgh4a     29998   ./b-file.txt
	// bafybeiaq6e55xratife7s5cmzjcmwy4adzzlk74sbdpfcq72gus6cweeeq     -       ./subdir
	// bafkreihdhgb5vyuqu7jssreyo3h567obewtqq37fi5hr2w4um5icacry7m     30999   ./subdir/leaf.txt
	root := f.MustGetNode().Cid().String()
	assert.Equal(t, "bafybeidlbwbu73tbjr3atntjz4lq5ego5w2uyof35vvwcnheaftzi3rndu", root)

	leaf := f.MustGetNode("subdir", "leaf.txt").Cid().String()
	assert.Equal(t, "bafkreihdhgb5vyuqu7jssreyo3h567obewtqq37fi5hr2w4um5icacry7m", leaf)

	nodes := f.MustGetChildren()

	assert.Len(t, nodes, 4)
	assert.Equal(t, "bafkreidw23elffhagxz3oi6ctoibqouzfowfn3bwcvq2yzgd5n5h4gjyou", nodes[0].Cid().String())
	assert.Equal(t, "bafkreiaeqsxxqwmsnhzhrlyr2udn25hpj24bs7gzcgkhbrkmhcuikcgh4a", nodes[1].Cid().String())
	assert.Equal(t, "bafybeiaq6e55xratife7s5cmzjcmwy4adzzlk74sbdpfcq72gus6cweeeq", nodes[2].Cid().String())
	assert.Equal(t, "bafkreihdhgb5vyuqu7jssreyo3h567obewtqq37fi5hr2w4um5icacry7m", nodes[3].Cid().String())

	cids := f.MustGetChildrenCids()
	assert.Len(t, nodes, 4)
	assert.Equal(t, "bafkreidw23elffhagxz3oi6ctoibqouzfowfn3bwcvq2yzgd5n5h4gjyou", cids[0])
	assert.Equal(t, "bafkreiaeqsxxqwmsnhzhrlyr2udn25hpj24bs7gzcgkhbrkmhcuikcgh4a", cids[1])
	assert.Equal(t, "bafybeiaq6e55xratife7s5cmzjcmwy4adzzlk74sbdpfcq72gus6cweeeq", cids[2])
	assert.Equal(t, "bafkreihdhgb5vyuqu7jssreyo3h567obewtqq37fi5hr2w4um5icacry7m", cids[3])
}
