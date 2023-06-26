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

	nodes := f.MustGetDescendants()

	assert.Len(t, nodes, 4)
	assert.Equal(t, "bafkreidw23elffhagxz3oi6ctoibqouzfowfn3bwcvq2yzgd5n5h4gjyou", nodes[0].Cid().String())
	assert.Equal(t, "bafkreiaeqsxxqwmsnhzhrlyr2udn25hpj24bs7gzcgkhbrkmhcuikcgh4a", nodes[1].Cid().String())
	assert.Equal(t, "bafybeiaq6e55xratife7s5cmzjcmwy4adzzlk74sbdpfcq72gus6cweeeq", nodes[2].Cid().String())
	assert.Equal(t, "bafkreihdhgb5vyuqu7jssreyo3h567obewtqq37fi5hr2w4um5icacry7m", nodes[3].Cid().String())

	cids := f.MustGetDescendantsCids()
	assert.Len(t, nodes, 4)
	assert.Equal(t, "bafkreidw23elffhagxz3oi6ctoibqouzfowfn3bwcvq2yzgd5n5h4gjyou", cids[0])
	assert.Equal(t, "bafkreiaeqsxxqwmsnhzhrlyr2udn25hpj24bs7gzcgkhbrkmhcuikcgh4a", cids[1])
	assert.Equal(t, "bafybeiaq6e55xratife7s5cmzjcmwy4adzzlk74sbdpfcq72gus6cweeeq", cids[2])
	assert.Equal(t, "bafkreihdhgb5vyuqu7jssreyo3h567obewtqq37fi5hr2w4um5icacry7m", cids[3])
}

func TestIssue54MustGetChildren(t *testing.T) {
	f := MustOpenUnixfsCar("./_fixtures/issue-54.car")

	// › npx ipfs-car ls ./issue-54.car --verbose
	// QmT5W42oo38Bi4mL2aktaMyqu7tqj7xWCuZhmoTfo7XtiU  -       .
	// QmUBzv8HDDtnivUvPGkqBmkCeMJKeAWhUZtb5D8ouGnATZ  -       ./sub1
	// QmZgfvZtoFdbJy4JmpPHc1NCXyA7Snim2L8e6zKspiUzhu  7       ./sub1/hello.txt
	// QmVtAZGRHTCzSNt1vRgz1UESvcc57ebEcDYTaDJjVu1SrA  -       ./sub2
	// Qmf4EqZZpFPcy6oKsc84dGS5EpPdXYZ1hq39Gemadu6hfW  7       ./sub2/hello.txt
	cids := f.MustGetDescendantsCids("sub1", "hello.txt")

	// › ipfs dag get QmZgfvZtoFdbJy4JmpPHc1NCXyA7Snim2L8e6zKspiUzhu | jq
	// {
	//   "Data": {
	//     "/": {
	//       "bytes": "CAIYByAFIAI"
	//     }
	//   },
	//   "Links": [
	//     {
	//       "Hash": {
	//         "/": "QmaATBg1yioWhYHhoA8XSUqD1Ya91KKCibWVD4USQXwaVZ"
	//       },
	//       "Name": "",
	//       "Tsize": 13
	//     },
	//     {
	//       "Hash": {
	//         "/": "QmdQEnYhrhgFKPCq5eKc7xb1k7rKyb3fGMitUPKvFAscVK"
	//       },
	//       "Name": "",
	//       "Tsize": 10
	//     }
	//   ]
	// }
	assert.Len(t, cids, 2)
	assert.Equal(t, "QmaATBg1yioWhYHhoA8XSUqD1Ya91KKCibWVD4USQXwaVZ", cids[0])
	assert.Equal(t, "QmdQEnYhrhgFKPCq5eKc7xb1k7rKyb3fGMitUPKvFAscVK", cids[1])

	// › ipfs dag get 'Qmb7KRN5qCAwTYXAdTd5JHzXXQv3BDRJQhcEuMJzdiGix6' | jq
	// {
	//   "Data": {
	//     "/": {
	//       "bytes": "CAE"
	//     }
	//   },
	//   "Links": [
	//     {
	//       "Hash": {
	//         "/": "QmZgfvZtoFdbJy4JmpPHc1NCXyA7Snim2L8e6zKspiUzhu"
	//       },
	//       "Name": "hello.txt",
	//       "Tsize": 117
	//     }
	//   ]
	// }
	cids = f.MustGetDescendantsCids("sub1")
	assert.Len(t, cids, 3)
	assert.Equal(t, "QmZgfvZtoFdbJy4JmpPHc1NCXyA7Snim2L8e6zKspiUzhu", cids[0])
	assert.Equal(t, "QmaATBg1yioWhYHhoA8XSUqD1Ya91KKCibWVD4USQXwaVZ", cids[1])
	assert.Equal(t, "QmdQEnYhrhgFKPCq5eKc7xb1k7rKyb3fGMitUPKvFAscVK", cids[2])
}

func TestMustGetChildrenDespiteMissingBlocks(t *testing.T) {
	f := MustOpenUnixfsCar("./_fixtures/file-3k-and-3-blocks-missing-block.car")

	// {
	//   "Data": {
	//     "/": {
	//       "bytes": "CAIYgBgggAgggAgggAg"
	//     }
	//   },
	//   "Links": [
	//     {
	//       "Hash": {
	//         "/": "QmXakb8wxp4Q9jysbKUDgEnWHXWCg3QEHTaHuhJdDndyN5"
	//       },
	//       "Name": "",
	//       "Tsize": 1035
	//     },
	//     {
	//       "Hash": {
	//         "/": "Qmed9q9vkn1KDh1NRPTxtUEZvbGouZkkQ5j5oLKtPpNJcf"
	//       },
	//       "Name": "",
	//       "Tsize": 1035
	//     },
	//     {
	//       "Hash": {
	//         "/": "QmSJX5xgXtnpYnAMbGZQg3YUBwDn75HMpAC6woxqpNgjD4"
	//       },
	//       "Name": "",
	//       "Tsize": 1035
	//     }
	//   ]
	// }

	// The block in the middle was filtered out
	assert.Equal(t, f.MustGetCid(), "QmZGmS2U9aD3EH8Vtea2fWpurjjry4rCJ2A2dSQFCCSFdB")

	// We should be able to retrieve children CID nonetheless
	assert.Equal(t, f.MustGetChildrenCids(), []string{
		"QmXakb8wxp4Q9jysbKUDgEnWHXWCg3QEHTaHuhJdDndyN5",
		"Qmed9q9vkn1KDh1NRPTxtUEZvbGouZkkQ5j5oLKtPpNJcf",
		"QmSJX5xgXtnpYnAMbGZQg3YUBwDn75HMpAC6woxqpNgjD4",
	})
}
