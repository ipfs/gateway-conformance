package main

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/ipfs/go-cid"
	carv2 "github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/blockstore"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
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
    return result.String()
}

func gen_data() []string {
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
		Version:  1,    // Usually '1'.
		Codec:    0x0129, // 0x71 means "dag-json" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhType:   0x13, // 0x20 means "sha2-512" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhLength: 64,   // sha2-512 hash has a 64-byte sum.
	}}

	// 4.1. Attempt to compute links with a intuitive-ish way
	links := []datamodel.Link{}
	for _, node := range nodes {
		link := lsys.MustComputeLink(lp, node)
		links = append(links, link)

		// This is awkward here:
		// I have serialized my link above I think, but I haven't encoded the 
		// payload, it happens below.
		// So how did we generated the link value?
		// I can encode my block using the code below but:
		// The MustComputeLink did the work, right? Why do it twice?
		// There is tight coupling between the link prototype definition in `lp` and the encoder used here, which is incomfortable
		// It's not clear if the encoding is the same, what if the mustcomputelink did something different here? what if there are 2 libraries with different versions, etc?
		// The protocol requires bytes to be equals, but I don't understand how to express this equality through code.
		dagjson.Encode(node, os.Stdout)
	}

	// 4.2. Attempt to compute links Trying to figure out the "right" way to do it.
	links2 := []datamodel.Link{}
	for _, node := range nodes {
		// Not clear what is the purpose of Store? Is it expected to always return a link?  What if I
		// pass in a 5GB block? Is it going to split it?
		lnk, err := lsys.Store(
			linking.LinkContext{},
			lp,
			node,
		)
		if err != nil {
			panic(err)
		}

		links2 = append(links2, lnk)
	}

	// 5.1. Attempt to load a link + block back. This will be needed for serialization
	node, err := lsys.Load(
		linking.LinkContext{},
		links2[0],
		basicnode.Prototype.Any,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(node) // got the node back, but not the block.

	// 5.2. Attempt to load a serialized block back. This will be needed for serialization
	fmt.Printf("%v, %v\n", store.Has(ctx, links2[0].String()), store.Has(ctx, links2[0].Binary()))
	fmt.Printf("%v\n", store.Get(ctx, links2[0].Binary()))

	// 6. Create the root node
	n, err := qp.BuildList(basicnode.Prototype.Any, 3, func(la datamodel.ListAssembler) {
		for _, link := range links2 {
			qp.ListEntry(la, qp.Link(link))
		}
	})
	if err != nil {
		panic(err)
	}

	// What do I do now? I have a node that links to many other nodes, but all I want is a list of blocks.
	// Do I need to lsys.store my node to get a block? I get back a link, but I don't want a link, I want a block.
	fmt.Println(n)
	dagjson.Encode(n, os.Stdout)
	return []string{}
}

func main() {
	blocks := gen_data()

	// Step 2: generate a car file.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()

	thisBlock := blocks.NewBlock([]byte("fish"))
	thatBlock := blocks.NewBlock([]byte("lobster"))
	andTheOtherBlock := blocks.NewBlock([]byte("barreleye"))

	tdir, err := os.MkdirTemp(os.TempDir(), "example-*")
	if err != nil {
		panic(err)
	}
	dst := filepath.Join(tdir, "sample-rw-bs-v2.car")
	roots := []cid.Cid{thisBlock.Cid(), thatBlock.Cid(), andTheOtherBlock.Cid()}

	rwbs, err := blockstore.OpenReadWrite(dst, roots, carv2.UseDataPadding(1413), carv2.UseIndexPadding(42))
	if err != nil {
		panic(err)
	}

	// Put all blocks onto the blockstore.
	blocks := []blocks.Block{thisBlock, thatBlock}
	if err := rwbs.PutMany(ctx, blocks); err != nil {
		panic(err)
	}
	fmt.Printf("Successfully wrote %v blocks into the blockstore.\n", len(blocks))

	// Finalize the blockstore to flush out the index and make a complete CARv2.
	if err := rwbs.Finalize(); err != nil {
		panic(err)
	}
}