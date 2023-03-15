package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-car/v2/blockstore"
)

/**
 * list all `*.car` file in the basePath directory, recursively
 */
func listAllCarFile(basePath string) []string {
	var carFiles []string

	filepath.WalkDir(basePath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".car" {
			path, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			carFiles = append(carFiles, path)
		}

		return nil
	})

	return carFiles
}

func MergeFixtures(outputPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	carFiles := listAllCarFile("./fixtures")

	// First list all the roots in our fixtures
	roots := make([]cid.Cid, 0)

	for _, f := range carFiles {
		fmt.Printf("processing %s\n", f)
		robs, err := blockstore.OpenReadOnly(f,
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
	rout, err := blockstore.OpenReadWrite(outputPath, roots)
	if err != nil {
		return err
	}

	// Then aggregate all our blocks.
	for _, f := range carFiles {
		fmt.Printf("processing %s\n", f)
		robs, err := blockstore.OpenReadOnly(f,
			blockstore.UseWholeCIDs(true),
		)
		if err != nil {
			return err
		}

		cids, err := robs.AllKeysChan(ctx)
		if err != nil {
			return err
		}

		for c := range cids {
			fmt.Printf("Adding %s\n", c.String())
			block, err := robs.Get(ctx, c)
			if err != nil {
				return err
			}

			rout.Put(ctx, block)
		}
	}

	fmt.Printf("Finalizing...\n")
	err = rout.Finalize()
	return err
}
