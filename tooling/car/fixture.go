package car

import (
	"context"
	"io"
	"strings"
	"time"

	files "github.com/ipfs/boxo/files"

	unixfile "github.com/ipfs/boxo/ipld/unixfs/file"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipld/go-ipld-prime"
	mb "github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"
)

type FixtureNode struct {
	node format.Node
	dsvc format.DAGService
}

func (n *FixtureNode) Cid() cid.Cid {
	return n.node.Cid()
}

// DNSSafeCidV1 returns the CID as a DNS-safe CIDv1 string suitable for use
// in subdomain gateway hostnames. If the underlying CID is v0, it is
// converted to v1 first. The encoding is chosen based on the codec:
// libp2p-key uses base36 (to fit Ed25519 keys within the 63-char DNS label
// limit), everything else uses base32.
func (n *FixtureNode) DNSSafeCidV1() string {
	c := n.Cid()
	if c.Version() == 0 {
		c = cid.NewCidV1(c.Type(), c.Hash())
	}
	var base mb.Encoding = mb.Base32
	if c.Type() == cid.Libp2pKey {
		base = mb.Base36
	}
	s, err := mb.Encode(base, c.Bytes())
	if err != nil {
		panic(err)
	}
	return s
}

func (n *FixtureNode) RawData() []byte {
	return n.node.RawData()
}

func (n *FixtureNode) Formatted(codecStr string) []byte {
	node := n.node.(ipld.Node)
	return FormatDagNode(node, codecStr)
}

func (n *FixtureNode) ToFile() files.File {
	f, err := unixfile.NewUnixfsFile(context.Background(), n.dsvc, n.node)
	if err != nil {
		panic(err)
	}

	r, ok := f.(files.File)

	if !ok {
		panic("not a file")
	}

	return r
}

func (n *FixtureNode) ReadFile() string {
	f := n.ToFile()

	buf := new(strings.Builder)
	_, err := io.Copy(buf, f)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

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
