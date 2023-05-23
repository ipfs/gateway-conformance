package car

// Note we force imports of dagcbor, dagjson, and other codecs below.
// They are registering themselves with the multicodec package
// during their `init()`.
import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/ipfs/boxo/blockservice"
	"github.com/ipfs/boxo/ipld/car/v2/blockstore"
	"github.com/ipfs/boxo/ipld/merkledag"
	"github.com/ipfs/boxo/ipld/unixfs/io"
	"github.com/ipfs/gateway-conformance/tooling/fixtures"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipld/go-ipld-prime"
	_ "github.com/ipld/go-ipld-prime/codec/cbor"
	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	_ "github.com/ipld/go-ipld-prime/codec/dagjson"
	_ "github.com/ipld/go-ipld-prime/codec/json"
	"github.com/ipld/go-ipld-prime/multicodec"
	mc "github.com/multiformats/go-multicodec"
)

type UnixfsDag struct {
	dsvc  format.DAGService
	cid   cid.Cid
	node  format.Node
	links map[string]*UnixfsDag
}

func newUnixfsDagFromCar(file string) (*UnixfsDag, error) {
	bs, err := blockstore.OpenReadOnly(file)
	if err != nil {
		return nil, err
	}
	bsvc := blockservice.New(bs, nil)
	dsvc := merkledag.NewDAGService(bsvc)
	root, err := bs.Roots()
	if err != nil {
		return nil, err
	}
	if len(root) != 1 {
		return nil, fmt.Errorf("expected 1 root, got %d", len(root))
	}
	return &UnixfsDag{dsvc: dsvc, cid: root[0]}, nil
}

func (d *UnixfsDag) loadLinks(node format.Node) (map[string]*UnixfsDag, error) {
	result := make(map[string]*UnixfsDag)
	dir, err := io.NewDirectoryFromNode(d.dsvc, node)
	if err != nil {
		return nil, err
	}
	links, err := dir.Links(context.Background())
	if err != nil {
		return nil, err
	}
	for _, l := range links {
		result[l.Name] = &UnixfsDag{dsvc: d.dsvc, cid: l.Cid}
	}

	return result, nil
}

func (d *UnixfsDag) getNode(names ...string) (format.Node, error) {
	for _, name := range names {
		node, err := d.getNode()
		if err != nil {
			return nil, err
		}

		if d.links == nil {
			d.links, err = d.loadLinks(node)
			if err != nil {
				return nil, err
			}
		}

		d = d.links[name]
		if d == nil {
			return nil, fmt.Errorf("no link named %s", strings.Join(names, "/"))
		}
	}

	if d.node == nil {
		node, err := d.dsvc.Get(context.Background(), d.cid)
		if err != nil {
			return nil, err
		}
		d.node = node
	}

	return d.node, nil
}

func (d *UnixfsDag) listChildren(names ...string) ([]*FixtureNode, error) {
	node, err := d.getNode(names...)
	if err != nil {
		return nil, err
	}

	result := []*FixtureNode{}
	var recursive func(node format.Node) error

	recursive = func(node format.Node) error {
		result = append(result, &FixtureNode{node: node, dsvc: d.dsvc})
		links := node.Links()

		for _, link := range links {
			node, err := link.GetNode(context.Background(), d.dsvc)
			if err != nil {
				return err
			}

			err = recursive(node)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err = recursive(node)
	if err != nil {
		return nil, err
	}

	return result[1:], nil
}

func (d *UnixfsDag) mustGetNode(names ...string) format.Node {
	node, err := d.getNode(names...)
	if err != nil {
		panic(err)
	}
	return node
}

func (d *UnixfsDag) MustGetNode(names ...string) *FixtureNode {
	return &FixtureNode{node: d.mustGetNode(names...), dsvc: d.dsvc}
}

func (d *UnixfsDag) MustGetChildren(names ...string) [](*FixtureNode) {
	nodes, err := d.listChildren(names...)
	if err != nil {
		panic(err)
	}
	return nodes
}

func (d *UnixfsDag) MustGetChildrenCids(names ...string) []string {
	nodes := d.MustGetChildren(names...)
	var cids []string
	for _, node := range nodes {
		cids = append(cids, node.Cid().String())
	}
	return cids
}

func (d *UnixfsDag) MustGetRoot() *FixtureNode {
	return d.MustGetNode()
}

func (d *UnixfsDag) MustGetCid(names ...string) string {
	return d.mustGetNode(names...).Cid().String()
}

func (d *UnixfsDag) MustGetRawData(names ...string) []byte {
	return d.mustGetNode(names...).RawData()
}

func (d *UnixfsDag) MustGetFormattedDagNode(codecStr string, names ...string) []byte {
	node := d.mustGetNode(names...).(ipld.Node)
	return FormatDagNode(node, codecStr)
}

func FormatDagNode(node ipld.Node, codecStr string) []byte {
	var codec mc.Code
	if err := codec.Set(codecStr); err != nil {
		panic(err)
	}

	encoder, err := multicodec.LookupEncoder(uint64(codec))
	if err != nil {
		panic(fmt.Errorf("invalid encoding: %s - %s", codec, err))
	}

	output := new(bytes.Buffer)

	err = encoder(node, output)

	if err != nil {
		panic(err)
	}

	return output.Bytes()
}

func MustOpenUnixfsCar(file string) *UnixfsDag {
	fixturePath := path.Join(fixtures.Dir(), file)

	if strings.HasPrefix(file, "./") {
		fixturePath = file
	}

	dag, err := newUnixfsDagFromCar(fixturePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return dag
}
