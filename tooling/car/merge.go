package car

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	carv2 "github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/blockstore"
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

	// Now prepare our new CAR file
	fmt.Printf("Opening the %s file, with roots: %v\n", outputPath, roots)
	options := []carv2.Option{blockstore.WriteAsCarV1(true)}
	rout, err := blockstore.OpenReadWrite(outputPath, roots[0:1], options...)
	if err != nil {
		return err
	}

	// Then aggregate all our blocks.
	for _, path := range inputPaths {
		fmt.Printf("processing %s\n", path)
		robs, err := blockstore.OpenReadOnly(path,
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
