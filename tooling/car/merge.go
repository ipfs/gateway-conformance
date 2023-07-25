package car

import (
	"context"
	"fmt"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	carv2 "github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/blockstore"
	"github.com/ipld/go-ipld-prime/fluent"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
	"github.com/ipld/go-ipld-prime/storage/memstore"
)

// https://github.com/ipld/go-ipld-prime/blob/65bfa53512f2328d19273e471ce4fd6d964055a2/storage/bsadapter/bsadapter.go#L111C1-L120C2
func cidFromBinString(key string) (cid.Cid, error) {
	l, k, err := cid.CidFromBytes([]byte(key))
	if err != nil {
		return cid.Undef, fmt.Errorf("bsrvadapter: key was not a cid: %w", err)
	}
	if l != len(key) {
		return cid.Undef, fmt.Errorf("bsrvadapter: key was not a cid: had %d bytes leftover", len(key)-l)
	}
	return k, nil
}

func Merge(inputPaths []string, outputPath string) error {
	// First list all the unique roots in our fixtures
	uniqRoots := make(map[string]cid.Cid)
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

		for _, root := range r {
			uniqRoots[root.String()] = root
		}
	}

	roots := make([]cid.Cid, 0)
	for _, root := range uniqRoots {
		roots = append(roots, root)
	}

	// Then aggregate all roots under a single one
	lsys := cidlink.DefaultLinkSystem()
	store := memstore.Store{Bag: make(map[string][]byte)}
	lsys.SetWriteStorage(&store)
	lsys.SetReadStorage(&store)

	// Adding to a map, they won't accept duplicate, hence the need for the uniqRoots
	node := fluent.MustBuildMap(basicnode.Prototype.Map, int64(len(roots)), func(ma fluent.MapAssembler) {
		ma.AssembleEntry("Links").CreateList(int64(len(roots)), func(na fluent.ListAssembler) {
			for _, root := range roots {
				na.AssembleValue().CreateMap(3, func(fma fluent.MapAssembler) {
					fma.AssembleEntry("Hash").AssignLink(cidlink.Link{Cid: root})
				})
			}
		})
	})

	lp := cidlink.LinkPrototype{Prefix: cid.Prefix{
		Version:  1,
		Codec:    0x70, // dag-pb
		MhType:   0x12,
		MhLength: 32, // sha2-256
	}}

	lnk, err := lsys.Store(
		linking.LinkContext{},
		lp,
		node)
	if err != nil {
		return err
	}

	rootCid := lnk.(cidlink.Link).Cid

	// Now prepare our new CAR file
	fmt.Printf("Opening the %s file, with root: %v\n", outputPath, rootCid)
	options := []carv2.Option{blockstore.WriteAsCarV1(true)}
	rout, err := blockstore.OpenReadWrite(outputPath, []cid.Cid{rootCid}, options...)
	if err != nil {
		fmt.Println("Error here:a", err)
		return err
	}

	// Add blocks from our store (root block)
	for k, v := range store.Bag {
		// cid.Parse and cid.Decode does not work here, using:
		// https://github.com/ipld/go-ipld-prime/blob/65bfa53512f2328d19273e471ce4fd6d964055a2/storage/bsadapter/bsadapter.go#L87-L89
		c, err := cidFromBinString(k)
		if err != nil {
			return err
		}

		blk, err := blocks.NewBlockWithCid(v, c)
		if err != nil {
			return err
		}

		err = rout.Put(context.Background(), blk)
		if err != nil {
			return err
		}
	}

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
