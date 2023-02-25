package car

import (
	"context"
	"fmt"
	"os"

	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-unixfs/io"
	"github.com/ipld/go-car/v2/blockstore"
)

type Directory struct {
	dsvc format.DAGService
	cid  cid.Cid
}

func (d *Directory) GetNode(names ...string) (format.Node, error) {
	for _, name := range names {
		node, err := d.GetNode()
		if err != nil {
			return nil, err
		}
		dir, err := io.NewDirectoryFromNode(d.dsvc, node)
		if err != nil {
			return nil, err
		}
		links, err := dir.Links(context.Background())
		if err != nil {
			return nil, err
		}
		var link *format.Link
		for _, l := range links {
			if l.Name == name {
				link = l
				break
			}
		}
		if link == nil {
			return nil, fmt.Errorf("no link named %s", name)
		}
		d = &Directory{d.dsvc, link.Cid}
	}
	return d.dsvc.Get(context.Background(), d.cid)
}

func NewDirectoryFromCar(file string) (*Directory, error) {
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
	return &Directory{dsvc, root[0]}, nil
}

func GetCid(car string, names ...string) string {
	dir, err := NewDirectoryFromCar(car)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	node, err := dir.GetNode(names...)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return node.Cid().String()
}

func GetRawData(car string, names ...string) []byte {
	dir, err := NewDirectoryFromCar(car)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	node, err := dir.GetNode(names...)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return node.RawData()
}
