package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/blockstore"
	"github.com/urfave/cli/v2"
)

func dropBlockFromCar(input, output, removedBlock string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Read roots & blocks
	// RELATED: https://github.com/ipld/go-car/issues/395
	robs, err := blockstore.OpenReadOnly(input, car.UseWholeCIDs(true))
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

	// Create trimmed car file, copy roots and blocks, except the removed one.
	rwbs, err := blockstore.OpenReadWrite(output, roots, car.UseWholeCIDs(true))
	if err != nil {
		panic(err)
	}

	found := false
	removedBlockCID := cid.MustParse(removedBlock)
	for _, block := range blocks {
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
		panic("removed block not found")
	}

	// Finalize the blockstore to flush out the index and make a complete CARv2.
	if err := rwbs.Finalize(); err != nil {
		panic(err)
	}

	fmt.Printf("Successfully removed %v blocks from the blockstore.\n", removedBlock)
	return nil
}

func main() {
	app := &cli.App{
		Name:  "deliberate-car",
		Usage: "Tooling for the gateway test suite",
		Commands: []*cli.Command{
			{
				Name:    "remove-block",
				Aliases: []string{"rb"},
				Usage:   "Remove block",
				Flags:   []cli.Flag{},
				Action: func(cCtx *cli.Context) error {
					input := cCtx.Args().Get(0)
					output := cCtx.Args().Get(1)
					removedBlock := cCtx.Args().Get(2)

					return dropBlockFromCar(input, output, removedBlock)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
