package car

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	format "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-unixfs/io"
	"github.com/ipld/go-car/v2/blockstore"
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
		fmt.Println(err)
		os.Exit(1)
	}
	return node
}

func (d *UnixfsDag) MustGetCid(names ...string) string {
	return d.mustGetNode(names...).Cid().String()
}

func (d *UnixfsDag) MustGetRawData(names ...string) []byte {
	return d.mustGetNode(names...).RawData()
}

func MustOpenUnixfsCar(file string) *UnixfsDag {
	dag, err := newUnixfsDagFromCar(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return dag
}
