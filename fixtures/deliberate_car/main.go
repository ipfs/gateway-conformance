package main

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	blocks "github.com/ipfs/go-block-format"
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
		Version:  1,    // Usually '1'.
		Codec:    0x0129, // 0x71 means "dag-json" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhType:   0x13, // 0x20 means "sha2-512" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhLength: 64,   // sha2-512 hash has a 64-byte sum.
	}}

	// 4.1. Attempt to compute links with a intuitive-ish way
	links := []datamodel.Link{}
	for _, node := range nodes {
		link := lsys.MustComputeLink(lp, node) // Does not store the block which makes it unusable, unclear why.
		links = append(links, link)
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
	lsys.SetReadStorage(&store)
	node, err := lsys.Load(
		linking.LinkContext{},
		links2[0],
		basicnode.Prototype.Any,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("node:", node) // got the node back, but not the block.

	// 5.2. Attempt to load a serialized block back. This will be needed for serialization
	ctx := context.Background()
	fmt.Printf("val link[0]: %v\n", links2[0].String())
	l, err := store.Has(ctx, links2[0].Binary())
	if err != nil {
		panic(err)
	}
	fmt.Printf("has link[0]: %v\n", l)

	bs, err := store.Get(ctx, links2[0].Binary())
	if err != nil {
		panic(err)
	}
	fmt.Printf("got link[0]: %s\n", bs)
	// that thing prints the string "random-xxxx", it seems to be the string payload of my node that was
	// the target of the link. It's not clear whether that's the block, the Node String, or some payload.
	// Looking at the output, that's the dag-json encoding of my link, which might be a block or the node.

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
	// This line will dump the dag-json encoding of my root node, which is a list of links.
	dagjson.Encode(n, os.Stdout) // If I uncomment this line, the code will panic because our global side effect disappear.
	fmt.Println()

	// Generate the nodes
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
	for _, link := range links2 {
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

func main() {
	bs := gen_data()

	for _, b := range bs {
		fmt.Printf("block: %s\n", b)
	}

	// Step 2: generate a car file.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// No imports in the example: https://pkg.go.dev/github.com/ipld/go-car/v2/blockstore#example-OpenReadWrite
	// So vscode imports "github.com/ipfs/go-libipfs/blocks" by defaults, which breaks.
	// import is actually: github.com/ipfs/go-block-format
	bs2 := []blocks.Block{}
	for _, b := range bs {
		bs2 = append(bs2, blocks.NewBlock(b))
	}

	tdir, err := os.MkdirTemp(os.TempDir(), "deliberate-*")
	if err != nil {
		panic(err)
	}
	dst := filepath.Join(tdir, "deliberate-complete.car")
	roots := []cid.Cid{bs2[len(bs2) -1].Cid()}

	// Why do I have these values? What are they? What do they mean?
	rwbs, err := blockstore.OpenReadWrite(dst, roots, carv2.UseDataPadding(1413), carv2.UseIndexPadding(42))
	if err != nil {
		panic(err)
	}

	// Put all blocks onto the blockstore.
	if err := rwbs.PutMany(ctx, bs2); err != nil {
		panic(err)
	}
	fmt.Printf("Successfully wrote %v blocks into the blockstore.\n", len(bs2))

	// Finalize the blockstore to flush out the index and make a complete CARv2.
	if err := rwbs.Finalize(); err != nil {
		panic(err)
	}
}