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

	// NOTE: by default we generate super large CIDs (from the doc)
	// which are not compatible with dns.
	// lp := cidlink.LinkPrototype{Prefix: cid.Prefix{
	// 	Version:  1,    // Usually '1'.
	// 	Codec:    0x71, // 0x71 means "dag-cbor" -- See the multicodecs table: https://github.com/multiformats/multicodec/
	// 	MhType:   0x13, // 0x20 means "sha2-512" -- See the multicodecs table: https://github.com/multiformats/multicodec/
	// 	MhLength: 64,   // sha2-512 hash has a 64-byte sum.
	// }}
	// cid: bafyrgqhpkthtuhnrvrnzobebylknmj4ayxac2f3kfm7pxm5ywmhu65ztzuyz4mmrhwf4sjliwntozivctgwk6qxiquospjybg37o4aiyvzt64
	// So I switch to sah2-256:

	// lp := cidlink.LinkPrototype{Prefix: cid.Prefix{
	// 	Version:  1,
	// 	Codec:    0x71,
	// 	MhType:   0x12,  // sha2-256
	// 	MhLength: 32,
	// }}
	// Trying to publish that configuratino:
	/*
		curl --oauth2-bearer "$W3STORAGE_TOKEN" --data-binary @fixtures.car "https://api.web3.storage/car"
		{"code":"HTTP_ERROR","message":"protobuf: (PBNode) invalid wireType, expected 2, got 0"}%

		What is that wireType? Maybe we can only publish dag-pb, so let's try.
	*/

	lp := cidlink.LinkPrototype{Prefix: cid.Prefix{
		Version:  1,
		Codec:    0x70, // dag-pb
		MhType:   0x12,
		MhLength: 32, // sha2-256
	}}
	/**
	Trying to run with that code, no the code doesn't complete:
	2023/07/05 10:57:36 func called on wrong kind: "AssignNode" called on a dagpb.PBNode node (kind: list), but only makes sense on map
	*/

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
		fmt.Println("Error here0:", err)
		return err
	}

	// Add blocks from our store (root block)
	for k, v := range store.Bag {
		fmt.Println("Adding block", k)
		c, err := cid.Parse(k)
		if err != nil {
			fmt.Println("Error here:", c, err)
		}

		c, err = cid.Decode(k)
		if err != nil {
			fmt.Println("Error here:", c, err)
			// Adding block q@T:['-j+>Ow311=%hf좢B'6g
			// 2023/07/05 10:24:43 invalid cid: selected encoding not supported
			// return err
		}

		// blk, err := blocks.NewBlockWithCid(v, c)
		// if err != nil {
		// 	return err
		// }

		blk := blocks.NewBlock(v)
		rout.Put(context.Background(), blk)
	}

	// Then aggregate all our blocks.
	for _, path := range inputPaths {
		fmt.Printf("processing %s\n", path)
		robs, err := blockstore.OpenReadOnly(
			path,
			blockstore.UseWholeCIDs(true),
		)
		if err != nil {
			fmt.Println("Error here:1", err)
			return err
		}

		cids, err := robs.AllKeysChan(context.Background())
		if err != nil {
			fmt.Println("Error here:2", err)
			return err
		}

		for c := range cids {
			fmt.Printf("Adding %s\n", c.String())
			block, err := robs.Get(context.Background(), c)
			if err != nil {
				fmt.Println("Error here:3", err)
				return err
			}

			rout.Put(context.Background(), block)
		}
	}

	fmt.Printf("Finalizing...\n")
	err = rout.Finalize()

	return err
}
