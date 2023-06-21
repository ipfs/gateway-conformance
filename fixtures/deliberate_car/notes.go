package main

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"time"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-car/v2/blockstore"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/fluent/qp"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/storage/memstore"
)

func randomString(size int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var seedRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	var result bytes.Buffer
	for i := 0; i < size; i++ {
		randomIndex := seedRand.Intn(len(charset))
		result.WriteByte(charset[randomIndex])
	}
	return ("random-" + result.String())[0:size]
}

func gen_data() [][]byte {
	// 1. Generate some Data
	chunks := []string{
		randomString(100),
		randomString(100),
		randomString(100),
	}

	// 2. Turn my data into IPLD Nodes (strings)
	nodes := []datamodel.Node{}
	for _, chunk := range chunks {
		nodes = append(nodes, basicnode.NewString(chunk))
	}

	// 3. Setup the linking system
	lsys := cidlink.DefaultLinkSystem()
	store := memstore.Store{}
	lsys.SetWriteStorage(&store)

	lp := cidlink.LinkPrototype{Prefix: cid.Prefix{
		Version:  1,      // Usually '1'.
		Codec:    0x0129, // 0x71 means "dag-json" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhType:   0x13,   // 0x20 means "sha2-512" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhLength: 64,     // sha2-512 hash has a 64-byte sum.
	}}

	// 4.1. Attempt to compute links with a intuitive-ish way
	// links := []datamodel.Link{}
	// for _, node := range nodes {
	// 	link := lsys.MustComputeLink(lp, node) // Does not store, unclear why.
	// 	links = append(links, link)
	// }

	// 4.2. Attempt to compute links Trying to figure out the "right" way to do it.
	links := []datamodel.Link{}
	for _, node := range nodes {
		lnk, err := lsys.Store(
			linking.LinkContext{},
			lp,
			node,
		)
		if err != nil {
			panic(err)
		}

		links = append(links, lnk)
	}

	// 5.1. Attempt to load a link + block back. This will be needed for serialization
	lsys.SetReadStorage(&store)
	node, err := lsys.Load(
		linking.LinkContext{},
		links[0],
		basicnode.Prototype.Any,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("node:", node) // got the node back, but not the block.

	// 5.2. Attempt to load a serialized block back. This will be needed for serialization
	ctx := context.Background()
	fmt.Printf("val link[0]: %v\n", links[0].String())
	l, err := store.Has(ctx, links[0].Binary())
	if err != nil {
		panic(err)
	}
	fmt.Printf("has link[0]: %v\n", l)

	bs, err := store.Get(ctx, links[0].Binary())
	if err != nil {
		panic(err)
	}
	fmt.Printf("got link[0]: %s\n", bs)
	// that thing prints the string "random-xxxx", it seems to be the string payload of my node that was
	// the target of the link. It's not clear whether that's the block, the Node String, or some payload.
	// Looking at the output, that's the dag-json encoding of my link, which might be a block or the node.

	// 6. Create the root node
	n, err := qp.BuildList(basicnode.Prototype.Any, 3, func(la datamodel.ListAssembler) {
		for _, link := range links {
			qp.ListEntry(la, qp.Link(link))
		}
	})
	if err != nil {
		panic(err)
	}

	// Store the root node/block
	lnk, err := lsys.Store(
		linking.LinkContext{},
		lp,
		n,
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("root link: %v\n", lnk.String())

	// 7. Attempt to return all the blocks
	blocks := [][]byte{}
	for _, link := range links {
		bs, err := store.Get(ctx, link.Binary())
		if err != nil {
			panic(err)
		}

		blocks = append(blocks, bs)
	}

	bs, err = store.Get(ctx, lnk.Binary())
	if err != nil {
		panic(err)
	}

	blocks = append(blocks, bs)
	return blocks
}

func loadBack() {
	cidPrintCount := 4

	// robs, err := blockstore.OpenReadOnly("./deliberate-complete.car")
	robs, err := blockstore.OpenReadOnly("./fixtures/deliberate_car/file_3k.car")
	if err != nil {
		panic(err)
	}
	defer robs.Close()

	// Print root CIDs.
	roots, err := robs.Roots()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Contains %v root CID(s):\n", len(roots))
	for _, r := range roots {
		fmt.Printf("\t%v\n", r)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Print the raw data size for the first 5 CIDs in the CAR file.
	keysChan, err := robs.AllKeysChan(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("List of first %v CIDs and their raw data size:\n", cidPrintCount)
	i := 1
	for k := range keysChan {
		if i > cidPrintCount {
			cancel()
			break
		}
		size, err := robs.GetSize(context.TODO(), k)
		if err != nil {
			panic(err)
		}
		fmt.Printf("\t%v -> %v bytes\n", k, size)
		i++
	}

}

func dropBlock() {
	input := "./file_3k.car"
	output := "./file_3k.trimmed.car"
	// removedBlock := "bafybeicrk7y5ub4pc4eoiidgwpyw7mgh34hxb2tmfh5za3xezrh2qutn24"
	removedBlock := "bafkreicrk7y5ub4pc4eoiidgwpyw7mgh34hxb2tmfh5za3xezrh2qutn24" // raw codec?!

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Read roots & blocks
	// RELATED: https://github.com/ipld/go-car/issues/395
	robs, err := blockstore.OpenReadOnly(input)
	if err != nil {
		panic(err)
	}
	defer robs.Close()

	roots, err := robs.Roots()
	if err != nil {
		panic(err)
	}

	keysChan, err := robs.AllKeysChan(ctx)
	if err != nil {
		panic(err)
	}

	blocks := make([]blocks.Block, 0)
	for k := range keysChan {
		bs, err := robs.Get(ctx, k)
		if err != nil {
			panic(err)
		}
		blocks = append(blocks, bs)
		fmt.Printf("block: key=%v; retrieved block cid=%v\n", k, bs.Cid().String())
	}
	robs.Close()

	// Update car file, remove block in place.
	found := false
	removedBlockCID := cid.MustParse(removedBlock)

	rwbs, err := blockstore.OpenReadWrite(output, roots)
	if err != nil {
		panic(err)
	}

	for _, block := range blocks {
		// That code won't work if you pass in a v0 CID (coming from ipfs dag get), because the CID in the car file are v1s.
		if block.Cid().Equals(removedBlockCID) {
			found = true
			// continue
		}

		err := rwbs.Put(ctx, block)
		if err != nil {
			panic(err)
		}
	}

	if !found {
		panic("block not found")
	}

	fmt.Printf("Successfully removed %v blocks from the blockstore.\n", removedBlock)

	// Finalize the blockstore to flush out the index and make a complete CARv2.
	if err := rwbs.Finalize(); err != nil {
		panic(err)
	}

	// loadBack()
}
