package car

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	carv2 "github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/blockstore"
	"github.com/ipld/go-ipld-prime/fluent"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/storage/memstore"
)

func Merge(inputPaths []string, outputPath string) error {
	// First list all the roots in our fixtures
	roots := make([]cid.Cid, 0)

	for _, path := range inputPaths {
		fmt.Printf("processing %s\n", path)
		robs, err := blockstore.OpenReadOnly(path,
			blockstore.UseWholeCIDs(true),
		)
		if err != nil {
			return err
		}

		r, err := robs.Roots()
		if err != nil {
			return err
		}

		roots = append(roots, r...)
	}

	// Then aggregate all roots under a single one
	lsys := cidlink.DefaultLinkSystem()
	store := memstore.Store{Bag: make(map[string][]byte)}
	lsys.SetWriteStorage(&store)
	lsys.SetReadStorage(&store)

	node := fluent.MustBuildList(basicnode.Prototype.List, int64(len(roots)), func(na fluent.ListAssembler) {
		for _, root := range roots {
			na.AssembleValue().AssignLink(cidlink.Link{Cid: root})
		}
	})

	lp := cidlink.LinkPrototype{Prefix: cid.Prefix{
		Version:  1,    // Usually '1'.
		Codec:    0x71, // 0x71 means "dag-cbor" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhType:   0x13, // 0x20 means "sha2-512" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhLength: 64,   // sha2-512 hash has a 64-byte sum.
	}}

	lnk, err := lsys.Store(
		linking.LinkContext{},
		lp,
		node)
	if err != nil {
		return err
	}

	rootCid := lnk.(cidlink.Link).Cid

	fmt.Printf("Root CID: %s\n", rootCid.String())

	// Now prepare our new CAR file
	fmt.Printf("Opening the %s file, with roots: %v\n", outputPath, roots)
	options := []carv2.Option{blockstore.WriteAsCarV1(true)}
	rout, err := blockstore.OpenReadWrite(outputPath, []cid.Cid{rootCid}, options...)
	if err != nil {
		return err
	}

	// Add blocks from our store (root block)
	// TODO: how to?
	// for every block in our store, add it to `rout`
	// for ever k, v in store.Bag ????

	// Then aggregate all our blocks.
	for _, path := range inputPaths {
		fmt.Printf("processing %s\n", path)
		robs, err := blockstore.OpenReadOnly(
			path,
			blockstore.UseWholeCIDs(true),
		)
		if err != nil {
			return err
		}

		cids, err := robs.AllKeysChan(context.Background())
		if err != nil {
			return err
		}

		for c := range cids {
			fmt.Printf("Adding %s\n", c.String())
			block, err := robs.Get(context.Background(), c)
			if err != nil {
				return err
			}

			rout.Put(context.Background(), block)
		}
	}

	fmt.Printf("Finalizing...\n")
	err = rout.Finalize()
	return err
}
