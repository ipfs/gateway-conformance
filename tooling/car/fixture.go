package car

import (
	"time"

	_ "github.com/ipfs/boxo/ipld/merkledag"

	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipld/go-ipld-prime"
	"github.com/multiformats/go-multihash"
)

type FixtureNode struct {
	node format.Node
	dsvc format.DAGService
}

// get cid method
func (n *FixtureNode) Cid() cid.Cid {
	return n.node.Cid()
}

// get raw data method
func (n *FixtureNode) RawData() []byte {
	return n.node.RawData()
}

// get formated node (pass codec name as parameter)
func (n *FixtureNode) Formatted(codecStr string) []byte {
	node := n.node.(ipld.Node)
	return FormatDagNode(node, codecStr)
}


// func (n *FixtureNode) ToFile() files.File {
//        f, err := unixfile.NewUnixfsFile(context.Background(), n.dsvc, n.node)
//        if err != nil {
//                panic(err)
//        }

//        r, ok := f.(files.File)

//        if !ok {
//                panic("not a file")
//        }

//        return r
// }

// // read the file
// func (n *FixtureNode) ReadFile() string {
//        f := n.ToFile()

//        buf := new(strings.Builder)
//        _, err := io.Copy(buf, f)
//        if err != nil {
//                panic(err)
//        }

//        return buf.String()
// }

func RandomCID() cid.Cid {
    now := time.Now().UTC()
    timeBytes := []byte(now.Format(time.RFC3339))

    mh, err := multihash.Sum(timeBytes, multihash.SHA2_256, -1)
    if err != nil {
        panic(err)
    }

    c := cid.NewCidV1(cid.Raw, mh)

	return c
}