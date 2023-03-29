package car

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
	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipld/go-ipld-prime"
	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
	_ "github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/multicodec"
	mc "github.com/multiformats/go-multicodec"
)

func init() {
	// Note we force imports of dagcbor and dagjson above.
	// They are registering themselves with the multicodec package
	// during their `init()`.
}

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

func (d *UnixfsDag) getNode(names ...string) (format.Node, error) {
	for _, name := range names {
		node, err := d.getNode()
		if err != nil {
			return nil, err
		}
		if d.links == nil {
			d.links = make(map[string]*UnixfsDag)
			dir, err := io.NewDirectoryFromNode(d.dsvc, node)
			if err != nil {
				return nil, err
			}
			links, err := dir.Links(context.Background())
			if err != nil {
				return nil, err
			}
			for _, l := range links {
				d.links[l.Name] = &UnixfsDag{dsvc: d.dsvc, cid: l.Cid}
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

func (d *UnixfsDag) mustGetNode(names ...string) format.Node {
	node, err := d.getNode(names...)
	if err != nil {
		panic(err)
	}
	return node
}

func (d *UnixfsDag) MustGetCid(names ...string) string {
	return d.mustGetNode(names...).Cid().String()
}

func (d *UnixfsDag) MustGetRawData(names ...string) []byte {
	return d.mustGetNode(names...).RawData()
}

func (d *UnixfsDag) MustGetFormattedDagNode(codecStr string, names ...string) []byte {
	var codec mc.Code
	if err := codec.Set(codecStr); err != nil {
		panic(err)
	}

	encoder, err := multicodec.LookupEncoder(uint64(codec))
	if err != nil {
		panic(fmt.Errorf("invalid encoding: %s - %s", codec, err))
	}

	output := new(bytes.Buffer)
	node := d.mustGetNode(names...).(ipld.Node)

	err = encoder(node, output)

	if err != nil {
		panic(err)
	}

	return output.Bytes()
}

func MustOpenUnixfsCar(file string) *UnixfsDag {
	fixturePath := path.Join(fixtures.Dir(), file)

	dag, err := newUnixfsDagFromCar(fixturePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return dag
}

func newBlockFromCar(file string) (blocks.Block, error) {
	bs, err := blockstore.OpenReadOnly(file)
	if err != nil {
		return nil, err
	}
	root, err := bs.Roots()
	if err != nil {
		return nil, err
	}
	if len(root) != 1 {
		return nil, fmt.Errorf("expected 1 root, got %d", len(root))
	}
	return bs.Get(context.Background(), root[0])
}

func MustOpenRawBlockFromCar(file string) blocks.Block {
	fixturePath := path.Join(fixtures.Dir(), file)

	block, err := newBlockFromCar(fixturePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return block
}
